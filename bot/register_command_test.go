package bot

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/google/go-cmp/cmp"
	"github.com/notion-echo/adapters/ent"
)

func TestRegisterCommandExecute(t *testing.T) {
	type fields struct {
		update *objects.Update
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
			"register user",
			fields{
				update: update(withMessage("/register"), withId(1)),
				bot:    bot(),
			},
			[]string{
				"click on the following URL, authorize pages",
				"localhost&state=stateToken",
				"when you have done with registration, select a default page using command `/defaultpage page`",
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

			sort.Strings(b.Resp)
			sort.Strings(tt.want)

			if diff := cmp.Diff(b.Resp, tt.want); diff != "" {
				t.Errorf("error %s: (- got, + want) %s\n", tt.name, diff)
			}
		})
	}
}
