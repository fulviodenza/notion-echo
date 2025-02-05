package bot

import (
	"context"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/r2"
	"github.com/notion-echo/bot/types"
	"github.com/sirupsen/logrus"
)

var _ types.IBot = (*MockBot)(nil)

type MockBot struct {
	Resp    []string
	Err     error
	usersDb db.UserRepoInterface
}

func NewMockBot(usersDb db.UserRepoInterface) *MockBot {
	return &MockBot{
		usersDb: usersDb,
	}
}

func (b *MockBot) SendMessage(msg string, chatId int, formatMarkdown bool, escape bool) error {
	if b.Err != nil {
		return b.Err
	}
	b.Resp = append(b.Resp, msg)
	return nil
}

func (b *MockBot) Start(ctx context.Context) {}

func (b *MockBot) GetHelpMessage() string {
	return "help message"
}

func (b *MockBot) SetTelegramClient(bot *tgbotapi.BotAPI) {}

func (b *MockBot) SetR2Client(bot r2.R2Interface) {}

func (b *MockBot) GetTelegramClient() *tgbotapi.BotAPI {
	return nil
}

func (b *MockBot) SetUserRepo(db db.UserRepoInterface) {

}
func (b *MockBot) GetUserRepo() db.UserRepoInterface {
	return b.usersDb
}

var (
	update = func(opts ...func(*tgbotapi.Update)) *tgbotapi.Update {
		update := &tgbotapi.Update{
			Message: &tgbotapi.Message{
				Chat: tgbotapi.Chat{
					ID: 1,
				},
			},
		}
		for _, o := range opts {
			o(update)
		}

		return update
	}
	withMessage = func(msg string) func(*tgbotapi.Update) {
		return func(up *tgbotapi.Update) {
			up.Message.Text = msg
		}
	}
	withId = func(id int) func(*tgbotapi.Update) {
		return func(up *tgbotapi.Update) {
			up.Message.Chat.ID = int64(id)
		}
	}
)

var (
	bot = func(opts ...func(*MockBot)) *MockBot {
		bot := NewMockBot(db.NewUserRepoMock(map[int]*ent.User{}, nil))
		for _, o := range opts {
			o(bot)
		}
		return bot
	}
	withUserRepo = func(repo db.UserRepoInterface) func(*MockBot) {
		return func(mb *MockBot) {
			mb.usersDb = repo
		}
	}
)

func (b *MockBot) SetNotionClient(token string, notionToken string) {}
func (b *MockBot) GetNotionClient(userId string) string             { return "" }
func (b *MockBot) SetNotionUser(token string)                       {}
func (b *MockBot) Logger() *logrus.Logger                           { return logrus.New() }
func (b *MockBot) GetUserState(userID int) string                   { return "" }
func (b *MockBot) SetUserState(userID int, msg string)              {}
func (b *MockBot) DeleteUserState(userID int)                       {}
