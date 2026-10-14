package bot

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/mymmrac/telego"
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

func (b *MockBot) SendButtonWithData(chatId int64, buttonText string, pages []*notionapi.Page) error {
	return nil
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

func (b *MockBot) SetTelegramClient(bot *telego.Bot) {}

func (b *MockBot) SetR2Client(bot r2.R2Interface) {}

func (b *MockBot) GetTelegramClient() *telego.Bot {
	return nil
}

func (b *MockBot) SetUserRepo(db db.UserRepoInterface) {

}
func (b *MockBot) GetUserRepo() db.UserRepoInterface {
	return b.usersDb
}

func (b *MockBot) SendButtonWithURL(chatId int64, buttonText, url, msgTxt string) error {
	return nil
}

var (
	update = func(opts ...func(*telego.Update)) *telego.Update {
		update := &telego.Update{
			Message: &telego.Message{
				Chat: telego.Chat{
					ID: 1,
				},
			},
		}
		for _, o := range opts {
			o(update)
		}

		return update
	}
	withMessage = func(msg string) func(*telego.Update) {
		return func(up *telego.Update) {
			up.Message.Text = msg
		}
	}
	withId = func(id int) func(*telego.Update) {
		return func(up *telego.Update) {
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
