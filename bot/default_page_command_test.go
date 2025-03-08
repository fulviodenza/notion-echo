package bot

import (
	"context"
	"errors"
	"os"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
)

func TestDefaultPageCommandExecute(t *testing.T) {
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

			t.Cleanup(func() {
				for k := range tt.fields.envs {
					os.Setenv(k, "")
				}
			})
		})
	}
}
