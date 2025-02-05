package bot

import (
	"context"
	"errors"
	"os"
	"sort"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	notionerrors "github.com/notion-echo/errors"
)

func TestDefaultPageCommandExecute(t *testing.T) {
	successResp := "page test set as default"
	selectPageResp := "write the page name in the next message"

	type fields struct {
		update               *tgbotapi.Update
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
				envs:   map[string]string{},
				bot:    bot(),
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
				bot:    bot(),
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
				envs:                 map[string]string{},
				bot:                  bot(),
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
				envs:   map[string]string{},
				bot:    bot(),
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
				envs:   map[string]string{},
				bot:    bot(),
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
				envs:   map[string]string{},
				bot:    bot(),
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
			"error getting default page",
			fields{
				update: update(withMessage("/defaultpage test"), withId(1)),
				envs:   map[string]string{},
				bot:    bot(),
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

			ec := NewDefaultPageCommand(tt.fields.bot, func(ctx context.Context, userRepo db.UserRepoInterface, id int, notionToken string) (notion.NotionInterface, error) {
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
