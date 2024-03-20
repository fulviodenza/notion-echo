package bot

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/notion-echo/parser"

	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/utils"

	bt "github.com/SakoDroid/telego/v2"
	cfg "github.com/SakoDroid/telego/v2/configs"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	notion "github.com/notion-echo/adapters/notion"
)

const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	NOTION_DATABASE_ID = "NOTION_DATABASE_ID"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	DATABASE_URL       = "DATABASE_URL"
	MAX_LEN_MESSAGE    = 4096
)

const (
	CATEGORIES_ASSET   = "./assets/categories.txt"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

const (
	COMMAND_NOTE = "/note"
	COMMAND_HELP = "/help"
)

const (
	PRIVATE_CHAT_TYPE    = "private"
	GROUP_CHAT_TYPE      = "group"
	SUPERGROUP_CHAT_TYPE = "supergroup"
)

type Bot struct {
	TelegramClient bt.Bot
	NotionClient   notion.Interface
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
	bot_config := &cfg.BotConfigs{
		BotAPI:         cfg.DefaultBotAPI,
		APIKey:         telegramToken,
		UpdateConfigs:  cfg.DefaultUpdateConfigs(),
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	bot := &Bot{}

	bot.loadHelpMessage()

	b, err := bt.NewBot(bot_config)
	if err != nil {
		return nil, err
	}
	bot.SetTelegramClient(*b)

	notionClient := notion.NewNotionService(notionapi.NewClient(notionapi.Token(notionToken)))
	bot.SetNotionClient(notionClient)

	userRepo, err := db.SetupAndConnectDatabase(databaseUrl)
	if err != nil {
		return nil, err
	}
	bot.SetUserRepo(userRepo)

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

func (b *Bot) GetHelpMessage() string {
	return b.helpMessage
}

func (b *Bot) SetNotionClient(client notion.Interface) {
	b.NotionClient = client
}

func (b *Bot) GetNotionClient() notion.Interface {
	return b.NotionClient
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
func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, up *objs.Update) {
	return map[string]func(ctx context.Context, up *objs.Update){
		COMMAND_NOTE: NewNoteCommand(b),
		COMMAND_HELP: NewHelpCommand(b),
	}
}
