package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/adapters/r2"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/oauth"
	"github.com/notion-echo/state"
	"github.com/notion-echo/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Bot struct {
	sync.RWMutex
	TelegramClient *tgbotapi.BotAPI
	NotionClient   map[string]string
	UserRepo       db.UserRepoInterface
	R2Client       r2.R2Interface
	helpMessage    string
	logger         *logrus.Logger
	state          state.IUserState
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ types.IBot = (*Bot)(nil)

var (
	telegramToken = os.Getenv(utils.TELEGRAM_TOKEN)
	databaseUrl   = os.Getenv(utils.DATABASE_URL)
	port          = os.Getenv(utils.PORT)
)

func NewBotWithConfig() (*Bot, error) {
	bot := &Bot{
		logger: logrus.New(),
	}
	bot.Logger().SetFormatter(&logrus.JSONFormatter{})

	bot.state = state.New()
	r2Client, err := r2.NewR2Client()
	if err != nil {
		fmt.Printf("got error: %v setting up r2 client", err)
	}
	logFileName := fmt.Sprintf("logs-%s.log", time.Now().Format("2006-01-02"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	bot.Logger().SetOutput(logFile)

	userRepo, err := db.SetupAndConnectDatabase(databaseUrl, bot.Logger())
	if err != nil {
		return nil, err
	}
	bot.SetUserRepo(userRepo)
	bot.loadHelpMessage()

	b, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, err
	}
	bot.SetTelegramClient(b)
	bot.SetR2Client(r2Client)

	// Schedule daily log upload
	go bot.scheduleDailyLogUpload(logFileName, r2Client.UploadLogs, bot.logger)

	return bot, err
}

func (b *Bot) scheduleDailyLogUpload(logFileName string, uploadFunc func(logFileName string, logger *logrus.Logger) error, logger *logrus.Logger) {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	durationUntilMidnight := nextMidnight.Sub(now)

	time.AfterFunc(durationUntilMidnight, func() {
		uploadFunc(logFileName, logger)

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			uploadFunc(logFileName, logger)
		}
	})
}

func (b *Bot) Start(ctx context.Context) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updateConfig.AllowedUpdates = []string{
		tgbotapi.UpdateTypeMessage,
		tgbotapi.UpdateTypeEditedMessage,
		tgbotapi.UpdateTypeChannelPost,
		tgbotapi.UpdateTypeEditedChannelPost,
		tgbotapi.UpdateTypeBusinessConnection,
		tgbotapi.UpdateTypeBusinessMessage,
		tgbotapi.UpdateTypeEditedBusinessMessage,
		tgbotapi.UpdateTypeDeletedBusinessMessages,
		tgbotapi.UpdateTypeMessageReaction,
		tgbotapi.UpdateTypeMessageReactionCount,
		tgbotapi.UpdateTypeInlineQuery,
		tgbotapi.UpdateTypeChosenInlineResult,
		tgbotapi.UpdateTypeCallbackQuery,
		tgbotapi.UpdateTypeShippingQuery,
		tgbotapi.UpdateTypePreCheckoutQuery,
		tgbotapi.UpdateTypePurchasedPaidMedia,
		tgbotapi.UpdateTypePoll,
		tgbotapi.UpdateTypePollAnswer,
		tgbotapi.UpdateTypeMyChatMember,
		tgbotapi.UpdateTypeChatMember,
		tgbotapi.UpdateTypeChatJoinRequest,
		tgbotapi.UpdateTypeChatBoost,
		tgbotapi.UpdateTypeRemovedChatBoost,
	}

	updatesChannel := b.TelegramClient.GetUpdatesChan(updateConfig)
	time.Sleep(time.Millisecond * 500)
	updatesChannel.Clear()

	b.Logger().Info("Bot started and waiting for updates...")

	for update := range updatesChannel {
		b.Logger().Info("Received an update")

		if update.Message == nil {
			continue
		}
		if strings.Contains(update.Message.Caption, "/note") {
			NewNoteCommand(b, buildNotionClient)(ctx, &update)
		}

		if state := b.GetUserState(int(update.Message.Chat.ID)); state != "" {
			switch state {
			case "/note":
				NewNoteCommand(b, buildNotionClient)(ctx, &update)
				b.state.Delete(int(update.Message.Chat.ID))
			case "/defaultpage":
				NewDefaultPageCommand(b, buildNotionClient)(ctx, &update)
				b.state.Delete(int(update.Message.Chat.ID))
			}
		}

		b.Logger().WithFields(logrus.Fields{"update_id": update.UpdateID}).Info("received update")

		var handlers = b.initializeHandlers()
		for c, f := range handlers {
			if strings.Contains(update.Message.Text, c) || strings.Contains(update.Message.Caption, c) {
				f(ctx, &update)
			}
		}
	}
}

func (b *Bot) RunOauth2Endpoint() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}}` + "\n",
	}))
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		b.Logger().Infof("[/] ok request")
		c.JSON(200, "ok")
		return nil
	})
	e.GET("/healthz", func(c echo.Context) error {
		b.Logger().Infof("[Healthz] healthz request")
		c.JSON(200, "ok")
		return nil
	})
	e.GET("/oauth2", func(c echo.Context) error {
		e.Logger.Info("received registration request")
		notionToken, err := oauth.Handler(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}

		state := c.QueryParam("state")
		_, err = b.GetUserRepo().SaveNotionTokenByStateToken(context.Background(), notionToken, state)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}
		b.SetNotionClient(state, notionToken)
		b.Logger().Info("[OAuth] user linked its notion")
		c.JSON(http.StatusOK, "your page has ben set, you can now close this page")
		return nil
	})

	// keyAuth := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
	// 	KeyLookup: fmt.Sprintf("header:%s", metricsAuthHeader),
	// 	Validator: func(key string, c echo.Context) (bool, error) {
	// 		if key == os.Getenv(metricsAuthHeader) {
	// 			return true, nil
	// 		}
	// 		return false, nil
	// 	},
	// })
	// e.GET("/metrics", echo.WrapHandler(promhttp.Handler()), keyAuth)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	address := fmt.Sprintf("0.0.0.0:%s", port)
	e.Logger.Fatal(e.Start(address))
}

func (b *Bot) GetHelpMessage() string {
	return b.helpMessage
}

func (b *Bot) SetNotionClient(token string, notionToken string) {
	b.Lock()
	defer b.Unlock()
	if b.NotionClient == nil {
		b.NotionClient = make(map[string]string)
	}
	b.NotionClient[token] = notionToken
}

func (b *Bot) GetNotionClient(userId string) string {
	b.RLock()
	defer b.RUnlock()
	return b.NotionClient[userId]
}

func (b *Bot) SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error {
	if escape {
		msg = utils.EscapeString(msg)
	}
	if len(msg) >= utils.MAX_LEN_MESSAGE {
		msgs := utils.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.Send(tgbotapi.NewMessage(int64(chatId), m))
			if err != nil {
				b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to send message")
				return err
			}
		}
	} else {
		_, err := b.TelegramClient.Send(tgbotapi.NewMessage(int64(chatId), msg))
		if err != nil {
			b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to send message")
			return err
		}
	}
	return nil
}

func (b *Bot) loadHelpMessage() {
	b.helpMessage = utils.HELP_STRING
}

func (b *Bot) SetTelegramClient(bot *tgbotapi.BotAPI) {
	b.TelegramClient = bot
}
func (b *Bot) GetTelegramClient() *tgbotapi.BotAPI {
	return b.TelegramClient
}
func (b *Bot) SetR2Client(bot r2.R2Interface) {
	b.R2Client = bot
}

func (b *Bot) SetUserRepo(db db.UserRepoInterface) {
	b.UserRepo = db
}
func (b *Bot) GetUserRepo() db.UserRepoInterface {
	return b.UserRepo
}

func (b *Bot) SetNotionUser(token string) {
	if b.NotionClient == nil {
		b.NotionClient = make(map[string]string)
	}
	b.NotionClient[token] = ""
}

func (b *Bot) Logger() *logrus.Logger {
	return b.logger
}

func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, update *tgbotapi.Update) {
	return map[string]func(ctx context.Context, update *tgbotapi.Update){
		"/start": func(ctx context.Context, update *tgbotapi.Update) {
			kb := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/note"),
					tgbotapi.NewKeyboardButton("/register"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/defaultpage"),
					tgbotapi.NewKeyboardButton("/getdefaultpage"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/help"),
				),
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to notion-echo bot!")
			msg.ReplyMarkup = kb
			_, err := b.TelegramClient.Send(msg)
			if err != nil {
				log.Println(err)
			}
		},
		utils.COMMAND_NOTE:             NewNoteCommand(b, buildNotionClient),
		utils.COMMAND_HELP:             NewHelpCommand(b),
		utils.COMMAND_REGISTER:         NewRegisterCommand(b, generateStateToken),
		utils.COMMAND_DEFAULT_PAGE:     NewDefaultPageCommand(b, buildNotionClient),
		utils.COMMAND_GET_DEFAULT_PAGE: NewGetDefaultPageCommand(b),
		utils.COMMAND_DEAUTHORIZE:      NewDeauthorizeCommand(b),
		// admin command
		utils.COMMAND_SEND_ALL: NewSendAllCommand(b),
	}
}

func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	stateToken := base64.URLEncoding.EncodeToString(b)
	return stateToken, nil
}

func buildNotionClient(ctx context.Context, userRepo db.UserRepoInterface, id int, notionToken string) (notion.NotionInterface, error) {
	token, err := userRepo.GetNotionTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return notion.NewNotionService(notionapi.NewClient(notionapi.Token(token))), nil
}

func (b *Bot) GetUserState(userID int) string {
	return b.state.Get(userID)
}

func (b *Bot) SetUserState(userID int, msg string) {
	b.state.Set(userID, msg)
}

func (b *Bot) DeleteUserState(userID int) {
	b.state.Delete(userID)
}
