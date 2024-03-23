package bot

import (
	"context"

	bt "github.com/SakoDroid/telego/v2"
	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/bot/types"
)

var _ types.IBot = (*MockBot)(nil)

type MockBot struct {
	Resp string
	Err  error
}

func NewMockBot() *MockBot {
	return &MockBot{}
}

func (b *MockBot) SendMessage(msg string, up *objects.Update, formatMarkdown bool) error {
	if b.Err != nil {
		return b.Err
	}
	b.Resp = msg
	return nil
}

func (b *MockBot) Start(ctx context.Context) {}

func (b *MockBot) GetHelpMessage() string {
	return ""
}

func (b *MockBot) SetTelegramClient(bot bt.Bot) {}

func (b *MockBot) GetTelegramClient() *bt.Bot {
	return nil
}

func (b *MockBot) SetUserRepo(db db.UserRepoInterface) {

}
func (b *MockBot) GetUserRepo() db.UserRepoInterface {
	return nil
}

var (
	update = func(opts ...func(*objects.Update)) *objects.Update {
		update := &objects.Update{
			Message: &objects.Message{
				Chat: &objects.Chat{
					Id: 1,
				},
			},
		}
		for _, o := range opts {
			o(update)
		}

		return update
	}
	withMessage = func(msg string) func(*objects.Update) {
		return func(up *objects.Update) {
			up.Message.Text = msg
		}
	}
)

var (
	bot = func(opts ...func(*MockBot)) *MockBot {
		bot := NewMockBot()
		for _, o := range opts {
			o(bot)
		}
		return bot
	}
)

func (b *MockBot) SetNotionClient(token string, notionToken string) {}
func (b *MockBot) GetNotionClient(userId string) string             { return "" }
func (b *MockBot) SetNotionUser(token string)                       {}
