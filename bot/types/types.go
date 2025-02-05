package types

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/r2"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

type NotionDbRow struct {
	Tags notionapi.MultiSelectProperty `json:"Tags"`
	Text notionapi.RichTextProperty    `json:"Text"`
	Name notionapi.TitleProperty       `json:"Name"`
}

type TelegramInterface interface {
	SendMessage(chatID int, msg string, parseMode string, replyToMessageID int, disableWebPagePreview, disableNotification bool) (*tgbotapi.Message, error)
}

type Command func(ctx context.Context, update *tgbotapi.Update)

type ICommand interface {
	Execute(ctx context.Context, update *tgbotapi.Update)
}

// Getters and Setters methods Bot instances
type IBot interface {
	Start(ctx context.Context)
	SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error
	GetHelpMessage() string
	SetTelegramClient(bot *tgbotapi.BotAPI)
	SetR2Client(r2 r2.R2Interface)
	GetTelegramClient() *tgbotapi.BotAPI
	SetNotionUser(token string)
	SetNotionClient(token string, notionToken string)
	GetNotionClient(userId string) string
	SetUserRepo(db db.UserRepoInterface)
	GetUserRepo() db.UserRepoInterface
	Logger() *logrus.Logger
	GetUserState(userID int) string
	SetUserState(userID int, msg string)
	DeleteUserState(userID int)
}
