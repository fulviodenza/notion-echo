package bot

import (
	"context"
	"errors"
	"os"
	"sort"
	"testing"

	"github.com/fulviodenza/telego/v2/objects"
	"github.com/google/go-cmp/cmp"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	notionerrors "github.com/notion-echo/errors"
)

func TestDefaultPageCommandExecute(t *testing.T) {
	successResp := "page test set as default"
	selectPageResp := "please, select a page"

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
			"set default page",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
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
			[]string{successResp},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"cannot find environment variables",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
				envs:   map[string]string{},
				bot:    bot(withVault("/localhost/test/", "testKey")),
			},
			[]string{"it looks like you are not registered, try running `/register` command first"},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"build notion client error",
			fields{
				update:               update(withMessage("/defaultpage test"), withId(1)),
				envs:                 map[string]string{"VAULT_PATH": "/localhost/test/"},
				bot:                  bot(withVault("/localhost/test/", "testKey")),
				buildNotionClientErr: errors.New(""),
			},
			[]string{notionerrors.ErrSetDefaultPage.Error()},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"empty page id received",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "",
						Object: notionapi.ObjectTypeBlock,
					},
				},
			},
			[]string{notionerrors.ErrPageNotFound.Error()},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"empty page object received",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "1",
						Object: "",
					},
				},
			},
			[]string{notionerrors.ErrPageNotFound.Error()},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"no page in command",
			fields{
				update: update(withMessage("/defaultpage"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey")),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "1",
						Object: "",
					},
				},
			},
			[]string{selectPageResp},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"multiple pages in command",
			fields{
				update: update(withMessage("/defaultpage test b"), withId(1)),
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
				"ignoring [b]",
				successResp,
			},
			&ent.User{
				ID: 1,
			},
			false,
		},
		{
			"error getting default page",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
				envs: map[string]string{
					"VAULT_PATH": "/localhost/test/",
				},
				bot: bot(withVault("/localhost/test/", "testKey"), withUserRepo(
					&db.UserRepoMock{
						Err: errors.New(""),
					},
				)),
				pages: map[string]*notionapi.Page{
					"test": {
						ID:     "1",
						Object: notionapi.ObjectTypeBlock,
					},
				},
			},
			[]string{
				notionerrors.ErrSetDefaultPage.Error(),
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

			ec := NewDefaultPageCommand(tt.fields.bot, func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error) {
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
