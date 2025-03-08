package bot

import (
	"context"
	"errors"
	"os"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
)

func TestRegisterCommandExecute(t *testing.T) {
	type fields struct {
		update *tgbotapi.Update
		bot    *MockBot
	}
	tests := []struct {
		name      string
		fields    fields
		wantUsers *ent.User
		err       bool
	}{
		{
			"register user",
			fields{
				update: update(withMessage("/register"), withId(1)),
				bot: bot(withUserRepo(&db.UserRepoMock{
					Db: map[int]*ent.User{
						1: {
							ID:          1,
							StateToken:  "token",
							DefaultPage: "test",
						},
					},
				})),
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"error registering user",
			fields{
				update: update(withMessage("/register"), withId(1)),
				bot: bot(withUserRepo(&db.UserRepoMock{
					Err: errors.New(""),
				})),
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OAUTH_URL", "localhost")
			b := tt.fields.bot
			ec := NewRegisterCommand(b, func() (string, error) {
				return "stateToken", nil
			})

			ec(context.Background(), tt.fields.update)

			if (b.Err != nil) != tt.err {
				t.Errorf("Bot.Execute() error = %v", b.Err)
			}
		})
	}
}
