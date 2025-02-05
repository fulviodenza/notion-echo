package bot

import (
	"context"
	"errors"
	"sort"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	notionerrors "github.com/notion-echo/errors"
)

func TestGetDefaultPageCommandExecute(t *testing.T) {
	type fields struct {
		update *tgbotapi.Update
		bot    *MockBot
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			"request default page",
			fields{
				update: update(withMessage("/getdefaultpage"), withId(1)),
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
			[]string{"your default page is **test**"},
		},
		{
			"empty page",
			fields{
				update: update(withMessage("/getdefaultpage"), withId(1)),
				bot: bot(withUserRepo(&db.UserRepoMock{
					Db: map[int]*ent.User{
						1: {
							ID:          1,
							StateToken:  "token",
							DefaultPage: "",
						},
					},
				})),
			},
			[]string{notionerrors.ErrPageNotFound.Error()},
		},
		{
			"error getting default page",
			fields{
				update: update(withMessage("/getdefaultpage"), withId(1)),
				bot: bot(withUserRepo(&db.UserRepoMock{
					Err: errors.New(""),
				})),
			},
			[]string{notionerrors.ErrPageNotFound.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.bot
			ec := NewGetDefaultPageCommand(b)

			ec(context.Background(), tt.fields.update)

			sort.Strings(b.Resp)
			sort.Strings(tt.want)

			if diff := cmp.Diff(b.Resp, tt.want); diff != "" {
				t.Errorf("error %s: (- got, + want) %s\n", tt.name, diff)
			}
		})
	}
}
