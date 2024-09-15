package types

import (
	"context"

	bt "github.com/SakoDroid/telego/v2"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/r2"
	"github.com/notion-echo/adapters/vault"
	"github.com/sirupsen/logrus"
)

type NotionDbRow struct {
	Tags notionapi.MultiSelectProperty `json:"Tags"`
	Text notionapi.RichTextProperty    `json:"Text"`
	Name notionapi.TitleProperty       `json:"Name"`
}

type TelegramInterface interface {
	SendMessage(chatID int, msg string, parseMode string, replyToMessageID int, disableWebPagePreview, disableNotification bool) (*objs.Message, error)
}

type Command func(ctx context.Context, update *objs.Update)

type ICommand interface {
	Execute(ctx context.Context, update *objs.Update)
}

// Getters and Setters methods Bot instances
type IBot interface {
	Start(ctx context.Context)
	SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error
	GetHelpMessage() string
	SetTelegramClient(bot bt.Bot)
	SetR2Client(r2 r2.R2Interface)
	GetTelegramClient() *bt.Bot
	SetNotionUser(token string)
	SetNotionClient(token string, notionToken string)
	GetNotionClient(userId string) string
	SetUserRepo(db db.UserRepoInterface)
	GetUserRepo() db.UserRepoInterface
	GetVaultClient() vault.VaultInterface
	SetVaultClient(v vault.VaultInterface)
	Logger() *logrus.Logger
	GetUserState(userID int) string
	SetUserState(userID int, msg string)
	DeleteUserState(userID int)
}
