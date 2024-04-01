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

	"github.com/notion-echo/adapters/db"
	vaultadapter "github.com/notion-echo/adapters/vault"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/oauth"
	"github.com/notion-echo/utils"

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
	VaultClient    vaultadapter.Vault
	helpMessage    string
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ types.IBot = (*Bot)(nil)

var (
	notionToken      = os.Getenv(utils.NOTION_TOKEN)
	notionDatabaseId = os.Getenv(utils.NOTION_DATABASE_ID)
	telegramToken    = os.Getenv(utils.TELEGRAM_TOKEN)
	databaseUrl      = os.Getenv(utils.DATABASE_URL)
	vaultSecretPath  = os.Getenv(utils.VAULT_PATH)
	vaultAddr        = os.Getenv(utils.VAULT_ADDR)
	vaultSecretKey   = os.Getenv(utils.VAULT_SECRET_KEY)
	vaultToken       = os.Getenv(utils.VAULT_TOKEN)
	port             = os.Getenv(utils.PORT)
)

func NewBotWithConfig() (*Bot, error) {
	bot := &Bot{}

	userRepo, err := db.SetupAndConnectDatabase(databaseUrl)
	if err != nil {
		return nil, err
	}
	bot.SetUserRepo(userRepo)

	vaultClient := vaultadapter.SetupVault(vaultAddr, vaultToken)
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
			log.Printf("got update: %v\n", update.Update_id)
		}
	}()

	var handlers = b.initializeHandlers()
	for c, f := range handlers {
		c := c
		f := f
		b.TelegramClient.AddHandler(c, func(u *objs.Update) {
			if strings.Contains(u.Message.Text, c) || strings.Contains(u.Message.Caption, c) {
				f(ctx, u)
			}
		}, utils.PRIVATE_CHAT_TYPE, utils.GROUP_CHAT_TYPE, utils.SUPERGROUP_CHAT_TYPE)
	}
}

func (b *Bot) RunOauth2Endpoint() {
	e := echo.New()
	e.Use(middleware.Logger())
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
		log.Println("received registration request")
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
		c.JSON(http.StatusOK, "ok")
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

func (b *Bot) SendMessage(msg string, up *objs.Update, formatMarkdown bool) error {
	parseMode := ""
	if formatMarkdown {
		parseMode = "Markdown"
	}

	if len(msg) >= utils.MAX_LEN_MESSAGE {
		msgs := utils.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, m, parseMode, 0, false, false)
			if err != nil {
				log.Printf("[SendMessage]: sending message to user: %v", err.Error())
				return err
			}
		}
	} else {
		_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, msg, parseMode, 0, false, false)
		if err != nil {
			log.Printf("[SendMessage]: sending message to user: %v", err.Error())
			return err
		}
	}
	return nil
}

func (b *Bot) loadHelpMessage() {
	helpMessage := make([]byte, 0)
	err := utils.Read(utils.HELP_MESSAGE_ASSET, &helpMessage)
	if err != nil {
		log.Fatalf("Failed to load help message: %v", err)
	}
	b.helpMessage = string(helpMessage)
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

func (b *Bot) SetVaultClient(v vaultadapter.Vault) {
	b.VaultClient = v
	encryptionKey, err := generateKey()
	if err != nil {
		log.Fatalf("Error generating key: %s", err)
	}
	encKeyStr := base64.StdEncoding.EncodeToString(encryptionKey)

	_, err = b.WriteOrGetSecret(vaultSecretPath, vaultSecretKey, encKeyStr)
	if err != nil {
		log.Fatalf("Error writing secret to Vault: %s", err)
	}
}
func (b *Bot) GetVaultClient() vaultadapter.Vault {
	return b.VaultClient
}

func (b *Bot) SetNotionUser(token string) {
	if b.NotionClient == nil {
		b.NotionClient = make(map[string]string)
	}
	b.NotionClient[token] = ""
}
func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, up *objs.Update) {
	return map[string]func(ctx context.Context, up *objs.Update){
		utils.COMMAND_NOTE:     NewNoteCommand(b),
		utils.COMMAND_HELP:     NewHelpCommand(b),
		utils.COMMAND_REGISTER: NewRegisterCommand(b),
		utils.COMMAND_START:    NewHelpCommand(b),
	}
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
		log.Printf("Failed to read secret: %v, assuming it does not exist and creating it.", err)

		data := map[string]interface{}{
			key: value,
		}
		_, err = b.VaultClient.Logical().Write(path, data)
		if err != nil {
			return "", fmt.Errorf("failed to write secret to Vault: %v", err)
		}
		log.Printf("Secret written to path %s.", path)
	} else {
		log.Printf("Secret at path %s already exists, not overwriting.", path)
	}

	return string(res), nil
}
