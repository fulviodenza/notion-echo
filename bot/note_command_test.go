package bot

import (
	"context"
	"os"
	"sort"
	"testing"

	"errors"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/google/go-cmp/cmp"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	boterrors "github.com/notion-echo/errors"
)

func TestNoteCommandExecute(t *testing.T) {
	type fields struct {
		update               *objects.Update
		envs                 map[string]string
		bot                  *MockBot
		pages                map[string]*notionapi.Page
		buildNotionClientErr error
	}
	tests := []struct {
		name      string
		fields    fields
		want      []string
		wantUsers *ent.User
		err       bool
	}{
		{
			"save note",
			fields{
				update: update(withMessage("/note test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "1",
						Object: notionapi.ObjectTypeBlock,
					},
				},
			},
			[]string{
				"note saved!",
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"save note",
			fields{
				update: update(withMessage("/note"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "1",
						Object: notionapi.ObjectTypeBlock,
					},
				},
			},
			[]string{
				"write something in your note!",
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"user not registered",
			fields{
				update: update(withMessage("/note test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot:                  bot(withVault("/localhost/test/", "testKey")),
				buildNotionClientErr: errors.New(""),
			},
			[]string{
				boterrors.ErrNotRegistered.Error(),
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"user notion page not found",
			fields{
				update: update(withMessage("/note test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
			},
			[]string{
				boterrors.ErrPageNotFound.Error(),
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.fields.envs {
				os.Setenv(k, v)
			}

			ec := NewNoteCommand(tt.fields.bot, func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error) {
				return notion.NewNotionMock(tt.fields.pages, tt.fields.bot.Err), tt.fields.buildNotionClientErr
			})

			ec(context.Background(), tt.fields.update)

			if (tt.fields.bot.Err != nil) != tt.err {
				t.Errorf("Bot.Execute() error = %v", tt.fields.bot.Err)
			}

			sort.Strings(tt.fields.bot.Resp)
			sort.Strings(tt.want)
			if diff := cmp.Diff(tt.fields.bot.Resp, tt.want); diff != "" {
				t.Errorf("error %s: (- got, + want) %s\n", tt.name, diff)
			}
			t.Cleanup(func() {
				for k := range tt.fields.envs {
					os.Setenv(k, "")
				}
			})
		})
	}
}
