package types

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/r2"
	"github.com/sirupsen/logrus"

	"github.com/mymmrac/telego"
)

type Command func(ctx context.Context, update *telego.Update)

type ICommand interface {
	Execute(ctx context.Context, update *telego.Update)
}

// Getters and Setters methods Bot instances
type IBot interface {
	Start(ctx context.Context)
	SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error
	SendButtonWithURL(chatId int64, buttonText, url, msgTxt string) error
	SendButtonWithData(chatId int64, buttonText string, pages []*notionapi.Page) error
	GetHelpMessage() string
	SetTelegramClient(bot *telego.Bot)
	SetR2Client(r2 r2.R2Interface)
	GetTelegramClient() *telego.Bot
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
