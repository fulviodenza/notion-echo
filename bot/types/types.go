package types

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/r2"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
)

type Command func(ctx context.Context, update *tgbotapi.Update)

type ICommand interface {
	Execute(ctx context.Context, update *tgbotapi.Update)
}

// Getters and Setters methods Bot instances
type IBot interface {
	Start(ctx context.Context)
	SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error
	SendButtonWithURL(chatId int64, buttonText, url, msgTxt string) error
	SendButtonWithData(chatId int64, buttonText string, pages []*notionapi.Page) error
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
