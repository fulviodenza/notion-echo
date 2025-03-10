package bot

import (
	"context"
	"sort"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
)

func TestDeauthorizeCommandExecute(t *testing.T) {
	type fields struct {
		update *tgbotapi.Update
		bot    *MockBot
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		err    bool
	}{
		{
			"deauthorize",
			fields{
				update: update(withMessage("/deauthorize"), withId(1)),
				bot:    bot(),
			},
			[]string{"deleted user"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.bot
			ec := NewDeauthorizeCommand(b)

			ec(context.Background(), tt.fields.update)

			sort.Strings(b.Resp)
			sort.Strings(tt.want)

			if diff := cmp.Diff(b.Resp, tt.want); diff != "" {
				t.Errorf("error %s: (- got, + want) %s\n", tt.name, diff)
			}

			u, err := tt.fields.bot.usersDb.GetUser(context.TODO(), int(tt.fields.update.Message.Chat.ID))
			if err != nil {
				t.Errorf("test: %v\nexpected user to not be present", tt.name)
			}
			if u != nil && !tt.err {
				t.Errorf("test: %v\nexpected user to not be present", tt.name)
			}
		})
	}
}
