package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	vaultadapter "github.com/notion-echo/adapters/vault"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/oauth"
	"github.com/notion-echo/utils"
	"github.com/sirupsen/logrus"

	bt "github.com/SakoDroid/telego/v2"
	cfg "github.com/SakoDroid/telego/v2/configs"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Bot struct {
	sync.RWMutex
	TelegramClient bt.Bot
	NotionClient   map[string]string
	UserRepo       db.UserRepoInterface
	VaultClient    vaultadapter.VaultInterface
	helpMessage    string
	logger         *logrus.Logger
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ types.IBot = (*Bot)(nil)

var (
	telegramToken   = os.Getenv(utils.TELEGRAM_TOKEN)
	databaseUrl     = os.Getenv(utils.DATABASE_URL)
	vaultSecretPath = os.Getenv(utils.VAULT_PATH)
	vaultAddr       = os.Getenv(utils.VAULT_ADDR)
	vaultSecretKey  = os.Getenv(utils.VAULT_SECRET_KEY)
	vaultToken      = os.Getenv(utils.VAULT_TOKEN)
	port            = os.Getenv(utils.PORT)
)

func NewBotWithConfig() (*Bot, error) {
	bot := &Bot{
		logger: logrus.New(),
	}
	bot.Logger().SetFormatter(&logrus.JSONFormatter{})

	userRepo, err := db.SetupAndConnectDatabase(databaseUrl, bot.Logger())
	if err != nil {
		return nil, err
	}
	bot.SetUserRepo(userRepo)

	vaultClient := vaultadapter.SetupVault(vaultAddr, vaultToken, bot.Logger())
	bot.SetVaultClient(vaultClient)

	bot.loadHelpMessage()

	bot_config := &cfg.BotConfigs{
		BotAPI:         cfg.DefaultBotAPI,
		APIKey:         telegramToken,
		UpdateConfigs:  cfg.DefaultUpdateConfigs(),
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}
	b, err := bt.NewBot(bot_config)
	if err != nil {
		return nil, err
	}
	bot.SetTelegramClient(*b)

	return bot, err
}

func (b *Bot) Start(ctx context.Context) {
	updateCh := b.TelegramClient.GetUpdateChannel()
	go func() {
		for {
			update := <-*updateCh
			if strings.Contains(update.Message.Caption, "/note") {
				NewNoteCommand(b, buildNotionClient)(ctx, update)
			}
			b.Logger().WithFields(logrus.Fields{"update_id": update.Update_id}).Info("received update")
		}
	}()

	var handlers = b.initializeHandlers()
	for c, f := range handlers {
		c := c
		f := f
		b.TelegramClient.AddHandler(c, func(u *objs.Update) {
			if strings.Contains(c, "/start") {
				kb := b.TelegramClient.CreateKeyboard(false, false, false, false, "type ...")

				kb.AddButton("/help", 1)
				kb.AddButton("/register", 1)
				kb.AddButton("/getdefaultpage", 2)

				_, err := b.TelegramClient.AdvancedMode().ASendMessage(u.Message.Chat.Id, "Welcome to notion-echo bot!", "", u.Message.MessageId, 0, false, false, nil, false, false, kb)
				if err != nil {
					fmt.Println(err)
				}
			}
			if strings.Contains(u.Message.Text, c) || strings.Contains(u.Message.Caption, c) {
				f(ctx, u)
			}
		}, utils.PRIVATE_CHAT_TYPE, utils.GROUP_CHAT_TYPE, utils.SUPERGROUP_CHAT_TYPE)
	}
}

func (b *Bot) RunOauth2Endpoint() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}", "method":"${method}", "uri":"${uri}", "status":${status}}` + "\n",
	}))
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		c.JSON(200, "ok")
		return nil
	})
	e.GET("/healthz", func(c echo.Context) error {
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
		encKey, err := b.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}
		notionTokenEnc, err := utils.EncryptString(notionToken, encKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}
		_, err = b.GetUserRepo().SaveNotionTokenByStateToken(context.Background(), notionTokenEnc, state)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return nil
		}
		b.SetNotionClient(state, notionToken)
		c.JSON(http.StatusOK, "your page has ben set, you can now close this page")
		return nil
	})
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

func (b *Bot) SendMessage(msg string, chatId int, formatMarkdown bool) error {
	msg = utils.EscapeString(msg)
	parseMode := ""
	if formatMarkdown {
		parseMode = "MarkdownV2"
	}

	if len(msg) >= utils.MAX_LEN_MESSAGE {
		msgs := utils.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(chatId, m, parseMode, 0, false, false)
			if err != nil {
				b.Logger().WithFields(logrus.Fields{"error": err}).Error("failed to send message")
				return err
			}
		}
	} else {
		_, err := b.TelegramClient.SendMessage(chatId, msg, parseMode, 0, false, false)
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

func (b *Bot) SetTelegramClient(bot bt.Bot) {
	b.TelegramClient = bot
}
func (b *Bot) GetTelegramClient() *bt.Bot {
	return &b.TelegramClient
}

func (b *Bot) SetUserRepo(db db.UserRepoInterface) {
	b.UserRepo = db
}
func (b *Bot) GetUserRepo() db.UserRepoInterface {
	return b.UserRepo
}

func (b *Bot) SetVaultClient(v vaultadapter.VaultInterface) {
	b.VaultClient = v
	encryptionKey, err := generateKey()
	if err != nil {
		b.Logger().WithFields(logrus.Fields{"error": err}).Fatal("error generating key")
	}
	encKeyStr := base64.StdEncoding.EncodeToString(encryptionKey)

	_, err = b.WriteOrGetSecret(vaultSecretPath, vaultSecretKey, encKeyStr)
	if err != nil {
		b.Logger().WithFields(logrus.Fields{"error": err}).Fatal("error writing secret to Vault")
	}
}
func (b *Bot) GetVaultClient() vaultadapter.VaultInterface {
	return b.VaultClient
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

func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, up *objs.Update) {
	return map[string]func(ctx context.Context, up *objs.Update){
		utils.COMMAND_NOTE:             NewNoteCommand(b, buildNotionClient),
		utils.COMMAND_HELP:             NewHelpCommand(b),
		utils.COMMAND_REGISTER:         NewRegisterCommand(b, generateStateToken),
		utils.COMMAND_START:            NewHelpCommand(b),
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

func generateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (b *Bot) WriteOrGetSecret(path string, key string, value string) (string, error) {
	res, err := b.VaultClient.GetKey(path)
	if err != nil || res == nil {
		b.Logger().Infof("failed to read secret: %v, assuming it does not exist and creating it.", err)

		data := map[string]interface{}{
			key: value,
		}
		_, err = b.VaultClient.Logical().Write(path, data)
		if err != nil {
			return "", fmt.Errorf("failed to write secret to Vault: %v", err)
		}
		b.Logger().Infof("secret written to path %s.", path)
	} else {
		b.Logger().Infof("secret at path %s already exists, not overwriting.", path)
	}

	return string(res), nil
}

func buildNotionClient(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error) {
	tokenEnc, err := userRepo.GetNotionTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}

	token, err := utils.DecryptString(tokenEnc, encKey)
	if err != nil {
		return nil, err
	}
	return notion.NewNotionService(notionapi.NewClient(notionapi.Token(token))), nil
}
