package bot

import (
	"context"
	"sort"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
	"github.com/notion-echo/adapters/ent"
)

func TestHelpCommandExecute(t *testing.T) {
	type fields struct {
		update *tgbotapi.Update
		bot    *MockBot
	}
	tests := []struct {
		name      string
		fields    fields
		want      []string
		wantUsers *ent.User
		err       bool
	}{
		{
			"help",
			fields{
				update: update(withMessage("/help"), withId(1)),
				bot:    bot(),
			},
			[]string{"help message"},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"help with parameter",
			fields{
				update: update(withMessage("/help a"), withId(1)),
				bot:    bot(),
			},
			[]string{"help message"},
			&ent.User{
				ID: 1,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.bot
			ec := NewHelpCommand(b)

			ec(context.Background(), tt.fields.update)

			if (b.Err != nil) != tt.err {
				t.Errorf("Bot.Execute() error = %v", b.Err)
			}

			sort.Strings(b.Resp)
			sort.Strings(tt.want)

			if diff := cmp.Diff(b.Resp, tt.want); diff != "" {
				t.Errorf("error %s: (- got, + want) %s\n", tt.name, diff)
			}
		})
	}
}
