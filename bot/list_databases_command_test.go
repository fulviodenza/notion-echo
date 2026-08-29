package bot

import (
	"context"
	"testing"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	boterrors "github.com/notion-echo/errors"
)

func TestListDatabasesCommandExecute(t *testing.T) {
	tests := []struct {
		name      string
		databases map[string]*notionapi.Database
		want      []string
	}{
		{
			name: "lists databases sorted",
			databases: map[string]*notionapi.Database{
				"finance": {ID: "1", Title: []notionapi.RichText{{Text: &notionapi.Text{Content: "Finance"}}}},
				"tasks":   {ID: "2", Title: []notionapi.RichText{{Text: &notionapi.Text{Content: "Tasks"}}}},
			},
			want: []string{"Databases I can access:\n• Finance\n• Tasks"},
		},
		{
			name:      "no databases",
			databases: map[string]*notionapi.Database{},
			want:      []string{boterrors.ErrDatabaseNotFound.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := notion.NewNotionMockWithDatabases(nil, tt.databases, nil)
			b := bot(withUserRepo(db.NewUserRepoMock(map[int]*ent.User{1: {ID: 1}}, nil)))

			cmd := &ListDatabasesCommand{
				IBot: b,
				buildNotionClient: func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error) {
					return mock, nil
				},
			}
			cmd.Execute(context.Background(), update(withMessage("/listdb"), withId(1)))

			if len(b.Resp) != len(tt.want) {
				t.Fatalf("responses = %v, want %v", b.Resp, tt.want)
			}
			for i := range tt.want {
				if b.Resp[i] != tt.want[i] {
					t.Errorf("response[%d] = %q, want %q", i, b.Resp[i], tt.want[i])
				}
			}
		})
	}
}
