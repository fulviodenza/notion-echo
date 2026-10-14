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
	"github.com/notion-echo/adapters/gladia"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/adapters/r2"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/oauth"
	"github.com/notion-echo/state"
	"github.com/notion-echo/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Bot struct {
	sync.RWMutex
	TelegramClient *telego.Bot
	NotionClient   map[string]string
	UserRepo       db.UserRepoInterface
	R2Client       r2.R2Interface
	helpMessage    string
	logger         *logrus.Logger
	state          state.IUserState
}

// this cast force us to follow the given interface
// if the interface will not be implemented, this will not compile
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

	b, err := telego.NewBot(telegramToken, telego.WithDefaultLogger(false, true))
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
	// Skip whatever piled up while the bot was down.
	if err := b.TelegramClient.DeleteWebhook(ctx, &telego.DeleteWebhookParams{DropPendingUpdates: true}); err != nil {
		b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to drop pending updates")
	}

	updatesChannel, err := b.TelegramClient.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        60,
		AllowedUpdates: []string{"message", "callback_query"},
	})
	if err != nil {
		b.Logger().WithFields(logrus.Fields{"error": err}).Fatal("failed to start long polling")
	}

	b.Logger().Info("Bot started and waiting for updates...")

	for update := range updatesChannel {
		b.Logger().Info("Received an update")
		if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
			if strings.HasPrefix(update.CallbackQuery.Data, "setpage:") {
				parts := strings.Split(update.CallbackQuery.Data, ":")
				if len(parts) != 2 {
					b.Logger().Error("invalid callback data format")
					continue
				}

				pageName := parts[1]
				chatID := update.CallbackQuery.Message.GetChat().ID

				err := b.GetUserRepo().SetDefaultPage(ctx, int(chatID), pageName)
				if err != nil {
					b.Logger().WithFields(logrus.Fields{
						"error":     err,
						"chat_id":   chatID,
						"page_name": pageName,
					}).Error("failed to set default page")

					b.SendMessage("Failed to set default page", int(chatID), false, true)
					continue
				}

				b.SendMessage(fmt.Sprintf("Default page set to: %s", pageName), int(chatID), false, true)

				b.TelegramClient.AnswerCallbackQuery(ctx, tu.CallbackQuery(update.CallbackQuery.ID).WithText("Default page updated"))
			}
		}

		if update.Message == nil {
			continue
		}
		if state := b.GetUserState(int(update.Message.Chat.ID)); state != "" {
			switch state {
			case "/note":
				NewNoteCommand(b, buildNotionClient)(ctx, &update)
				b.state.Delete(int(update.Message.Chat.ID))
				continue
			case "/defaultpage":
				NewDefaultPageCommand(b, buildNotionClient)(ctx, &update)
				b.state.Delete(int(update.Message.Chat.ID))
				continue
			}
		}

		if update.Message.Caption != "" && strings.Contains(update.Message.Caption, "/note") {
			NewNoteCommand(b, buildNotionClient)(ctx, &update)
			continue
		}

		if update.Message.Voice != nil {
			if update.Message.Voice.Duration > 30 {
				b.SendMessage("Voice notes must not last more than 30 seconds", int(update.Message.Chat.ID), false, true)
				continue
			}

			b.SendMessage(
				`Received transcription request, please wait until your transcription will be ready, a confirmation message will be sent`,
				int(update.Message.Chat.ID), false, true)
			message, err := gladia.HandleTranscribe(ctx, b.TelegramClient, update.Message.Voice)
			if err != nil {
				b.Logger().WithFields(logrus.Fields{
					"error": err,
				}).Error("Failed to transcribe voice message")
				b.SendMessage("Sorry, I couldn't transcribe your voice message.", int(update.Message.Chat.ID), false, true)
			} else {
				b.SendMessage(fmt.Sprintf("Transcribed: %s", message), int(update.Message.Chat.ID), false, true)

				noteUpdate := update
				noteUpdate.Message.Text = "/note " + message
				NewNoteCommand(b, buildNotionClient)(ctx, &noteUpdate)
			}
			continue
		}

		var foundCommand bool
		var handlers = b.initializeHandlers()
		for c, f := range handlers {
			if (update.Message.Text != "" && strings.HasPrefix(update.Message.Text, c)) ||
				(update.Message.Caption != "" && strings.HasPrefix(update.Message.Caption, c)) {
				f(ctx, &update)
				foundCommand = true
				break
			}
		}

		// If no command was found and there's text, treat it as a note
		if !foundCommand && update.Message.Text != "" {
			noteUpdate := update
			noteUpdate.Message.Text = "/note " + update.Message.Text
			NewNoteCommand(b, buildNotionClient)(ctx, &noteUpdate)
		}
	}
}

func (b *Bot) RunOauth2Endpoint() {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		HandleError: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("{\"time\":%q, \"method\":%q, \"uri\":%q, \"status\":%d}\n",
				v.StartTime.Format(time.RFC3339), v.Method, v.URI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/", func(c *echo.Context) error {
		b.Logger().Infof("[/] ok request")
		c.JSON(200, "ok")
		return nil
	})
	e.GET("/healthz", func(c *echo.Context) error {
		b.Logger().Infof("[Healthz] healthz request")
		c.JSON(200, "ok")
		return nil
	})
	e.GET("/oauth2", func(c *echo.Context) error {
		e.Logger.Info("received registration request")
		notionToken, err := oauth.Handler(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}

		state := c.QueryParam("state")

		fmt.Printf("got state token: %s", state)
		fmt.Printf("got notion token: %s", notionToken)
		_, err = b.GetUserRepo().SaveNotionTokenByStateToken(context.Background(), notionToken, state)
		if err != nil {
			fmt.Printf("got error saving notion token: %s, err: %v", notionToken, err)
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
	// 	Validator: func(key string, c *echo.Context) (bool, error) {
	// 		if key == os.Getenv(metricsAuthHeader) {
	// 			return true, nil
	// 		}
	// 		return false, nil
	// 	},
	// })
	// e.GET("/metrics", echo.WrapHandler(promhttp.Handler()), keyAuth)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	address := fmt.Sprintf("0.0.0.0:%s", port)
	log.Fatal(e.Start(address))
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
	fmt.Println(notionToken)
	b.NotionClient[token] = notionToken
}

func (b *Bot) GetNotionClient(userId string) string {
	b.RLock()
	defer b.RUnlock()
	return b.NotionClient[userId]
}

func (b *Bot) SendButtonWithURL(chatId int64, buttonText, url, msgTxt string) error {
	msg := tu.Message(tu.ID(chatId), msgTxt).WithReplyMarkup(tu.InlineKeyboard(
		tu.InlineKeyboardRow(tu.InlineKeyboardButton(buttonText).WithURL(url)),
	))

	_, err := b.TelegramClient.SendMessage(context.Background(), msg)
	if err != nil {
		b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to send message with button")
		return err
	}
	return nil
}

func (b *Bot) SendButtonWithData(chatId int64, buttonText string, pages []*notionapi.Page) error {
	var rows [][]telego.InlineKeyboardButton
	for _, page := range pages {
		title := notion.ExtractName(page.Properties)
		if title == "" {
			continue
		}
		// Use shorter callback data format: "sp:{pageID}"
		callbackData := fmt.Sprintf("setpage:%s", title)
		if len(callbackData) > 64 {
			b.Logger().WithFields(logrus.Fields{
				"page_id": page.ID,
				"title":   title,
			}).Warn("callback data too long, skipping page")
			continue
		}

		rows = append(rows, tu.InlineKeyboardRow(tu.InlineKeyboardButton(title).WithCallbackData(callbackData)))
	}

	if len(rows) == 0 {
		return fmt.Errorf("no valid pages to display")
	}

	msg := tu.Message(tu.ID(chatId), buttonText).WithReplyMarkup(tu.InlineKeyboard(rows...))

	_, err := b.TelegramClient.SendMessage(context.Background(), msg)
	return err
}

func (b *Bot) SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error {
	if len(msg) >= utils.MAX_LEN_MESSAGE {
		msgs := utils.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(context.Background(), tu.Message(tu.ID(int64(chatId)), m))
			if err != nil {
				b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to send message")
				return err
			}
		}
	} else {
		_, err := b.TelegramClient.SendMessage(context.Background(), tu.Message(tu.ID(int64(chatId)), msg))
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

func (b *Bot) SetTelegramClient(bot *telego.Bot) {
	b.TelegramClient = bot
}
func (b *Bot) GetTelegramClient() *telego.Bot {
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

func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, update *telego.Update) {
	return map[string]func(ctx context.Context, update *telego.Update){
		"/start": func(ctx context.Context, update *telego.Update) {
			kb := tu.Keyboard(
				tu.KeyboardRow(
					tu.KeyboardButton("/note"),
					tu.KeyboardButton("/register"),
				),
				tu.KeyboardRow(
					tu.KeyboardButton("/defaultpage"),
					tu.KeyboardButton("/getdefaultpage"),
				),
				tu.KeyboardRow(
					tu.KeyboardButton("/help"),
				),
			)

			msg := tu.Message(tu.ID(update.Message.Chat.ID), "Welcome to notion-echo bot!").WithReplyMarkup(kb)
			_, err := b.TelegramClient.SendMessage(ctx, msg)
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
		utils.COMMAND_LIST_DB:          NewListDatabasesCommand(b, buildNotionClient),
		utils.COMMAND_DB_NOTE:          NewDbNoteCommand(b, buildNotionClient),
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
