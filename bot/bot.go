package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/notion-echo/oauth"
	"github.com/notion-echo/parser"

	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/utils"

	bt "github.com/SakoDroid/telego/v2"
	cfg "github.com/SakoDroid/telego/v2/configs"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/adapters/db"
)

const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	NOTION_DATABASE_ID = "NOTION_DATABASE_ID"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	DATABASE_URL       = "DATABASE_URL"
	OAUTH_CLIENT_ID    = "OAUTH_CLIENT_ID"
	REDIRECT_URL       = "REDIRECT_URL"

	MAX_LEN_MESSAGE = 4096
)

const (
	CATEGORIES_ASSET   = "./assets/categories.txt"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

const (
	COMMAND_NOTE     = "/note"
	COMMAND_HELP     = "/help"
	COMMAND_REGISTER = "/register"
)

const (
	PRIVATE_CHAT_TYPE    = "private"
	GROUP_CHAT_TYPE      = "group"
	SUPERGROUP_CHAT_TYPE = "supergroup"
)

type Bot struct {
	sync.RWMutex
	TelegramClient bt.Bot
	NotionClient   map[string]string
	UserRepo       db.UserRepoInterface
	helpMessage    string
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ types.IBot = (*Bot)(nil)

// get variables from env
var (
	notionToken      = os.Getenv(NOTION_TOKEN)
	notionDatabaseId = os.Getenv(NOTION_DATABASE_ID)
	telegramToken    = os.Getenv(TELEGRAM_TOKEN)
	databaseUrl      = os.Getenv(DATABASE_URL)
)

func NewBotWithConfig() (*Bot, error) {
	bot := &Bot{}

	userRepo, err := db.SetupAndConnectDatabase(databaseUrl)
	if err != nil {
		return nil, err
	}
	bot.SetUserRepo(userRepo)

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
			if strings.Contains(u.Message.Text, c) {
				f(ctx, u)
			}
		}, PRIVATE_CHAT_TYPE, GROUP_CHAT_TYPE, SUPERGROUP_CHAT_TYPE)
	}
}

func (b *Bot) RunOauth2Endpoint() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		client, notionToken, token, err := oauth.Handler(c)
		state := c.QueryParam("state")
		fmt.Printf("state token: %s\n", state)
		b.UserRepo.SaveNotionTokenByStateToken(context.Background(), notionToken, state)
		b.SetNotionClient(token, notionToken)
		fmt.Println(b.NotionClient)
		c.JSON(200, client)
		return err
	})
	e.Logger.Fatal(e.StartTLS(":8080", "certs/cert.pem", "certs/key.pem"))
}

func (b *Bot) GetHelpMessage() string {
	return b.helpMessage
}

func (b *Bot) SetNotionClient(token string, notionToken string) {
	b.Lock()
	defer b.Unlock()
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

	if len(msg) >= MAX_LEN_MESSAGE {
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
	err := parser.Read(HELP_MESSAGE_ASSET, &helpMessage)
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

func (b *Bot) SetNotionUser(token string) {
	if b.NotionClient == nil {
		b.NotionClient = make(map[string]string)
	}
	b.NotionClient[token] = ""
}
func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, up *objs.Update) {
	return map[string]func(ctx context.Context, up *objs.Update){
		COMMAND_NOTE:     NewNoteCommand(b),
		COMMAND_HELP:     NewHelpCommand(b),
		COMMAND_REGISTER: NewRegisterCommand(b),
	}
}
