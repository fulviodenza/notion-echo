package bot

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/adapters/ent"
)

func TestExecute(t *testing.T) {
	successResponse := "localhost&state="
	type fields struct {
		update *objects.Update
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantUsers *ent.User
		err       bool
	}{
		{
			"register user",
			fields{
				update: update(withMessage("/register"), withId(1)),
			},
			successResponse,
			&ent.User{
				ID: 1,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OAUTH_URL", "localhost")
			b := bot()
			ec := NewRegisterCommand(b)

			ec(context.Background(), tt.fields.update)

			if (b.Err != nil) != tt.err {
				t.Errorf("Bot.Execute() error = %v", b.Err)
			}
			if !strings.Contains(b.Resp, tt.want) {
				t.Errorf("error %s: got: %s, want: %s\n", tt.name, b.Resp, tt.want)
			}

			if u, err := b.GetUserRepo().GetStateTokenById(context.TODO(), tt.fields.update.Message.Chat.Id); err != nil || u == nil {
				if u == nil && tt.err == false {
					t.Errorf("expected user to be present")
				}
			}
		})
	}
}
