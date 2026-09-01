package bot

import (
	"context"
	"testing"
	"time"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	boterrors "github.com/notion-echo/errors"
)

func financeDB() *notionapi.Database {
	return &notionapi.Database{
		ID:    "db-1",
		Title: []notionapi.RichText{{Text: &notionapi.Text{Content: "Finance"}}},
		Properties: notionapi.PropertyConfigs{
			"Name":   notionapi.TitlePropertyConfig{Type: notionapi.PropertyConfigTypeTitle},
			"Date":   notionapi.DatePropertyConfig{Type: notionapi.PropertyConfigTypeDate},
			"Status": notionapi.StatusPropertyConfig{Type: statusPropertyConfigType},
			"Tag":    notionapi.SelectPropertyConfig{Type: notionapi.PropertyConfigTypeSelect},
		},
	}
}

func namedDB(id, title string) *notionapi.Database {
	return &notionapi.Database{
		ID:    notionapi.ObjectID(id),
		Title: []notionapi.RichText{{Text: &notionapi.Text{Content: title}}},
		Properties: notionapi.PropertyConfigs{
			"Name": notionapi.TitlePropertyConfig{Type: notionapi.PropertyConfigTypeTitle},
		},
	}
}

func TestSelectDatabase(t *testing.T) {
	expenses := namedDB("db-exp", "silly little Expenses (database)")
	postits := namedDB("db-post", "silly little Post-its (database)")

	tests := []struct {
		name      string
		databases []*notionapi.Database
		query     string
		wantID    notionapi.ObjectID
		wantOk    bool
	}{
		{
			name:      "exact match wins over fuzzy first result",
			databases: []*notionapi.Database{postits, expenses},
			query:     "silly little Expenses (database)",
			wantID:    "db-exp",
			wantOk:    true,
		},
		{
			name:      "case-insensitive exact match",
			databases: []*notionapi.Database{postits, expenses},
			query:     "SILLY LITTLE expenses (database)",
			wantID:    "db-exp",
			wantOk:    true,
		},
		{
			name:      "single fuzzy result is used",
			databases: []*notionapi.Database{expenses},
			query:     "Expenses",
			wantID:    "db-exp",
			wantOk:    true,
		},
		{
			name:      "no exact match among several is rejected",
			databases: []*notionapi.Database{postits, expenses},
			query:     "Groceries",
			wantOk:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := selectDatabase(tt.databases, tt.query)
			if ok != tt.wantOk {
				t.Fatalf("selectDatabase() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && got.ID != tt.wantID {
				t.Errorf("selectDatabase() id = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}

func TestParseDbNoteMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    dbNoteRequest
		wantErr bool
	}{
		{
			name:    "db name and title",
			message: `/dbnote "Finance" coffee 3.50`,
			want:    dbNoteRequest{dbName: "Finance", title: "coffee 3.50"},
		},
		{
			name:    "today flag",
			message: `/dbnote "Finance" coffee 3.50 --today`,
			want:    dbNoteRequest{dbName: "Finance", title: "coffee 3.50", today: true},
		},
		{
			name:    "all flags, flag before title",
			message: `/dbnote "Finance" --status "Paid" --select "Food" --date "2026-08-29" lunch`,
			want:    dbNoteRequest{dbName: "Finance", title: "lunch", status: "Paid", sel: "Food", date: "2026-08-29"},
		},
		{
			name:    "em dash today flag",
			message: `/dbnote "Finance" coffee —today`,
			want:    dbNoteRequest{dbName: "Finance", title: "coffee", today: true},
		},
		{
			name:    "missing quotes",
			message: `/dbnote Finance coffee`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDbNoteMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDbNoteMessage() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("parseDbNoteMessage() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResolveDateProperty(t *testing.T) {
	tests := []struct {
		name  string
		props notionapi.PropertyConfigs
		want  string
		weOk  bool
	}{
		{
			name: "prefers common name over first alphabetical",
			props: notionapi.PropertyConfigs{
				"Appointment": notionapi.DatePropertyConfig{Type: notionapi.PropertyConfigTypeDate},
				"Created":     notionapi.DatePropertyConfig{Type: notionapi.PropertyConfigTypeDate},
			},
			want: "Created",
			weOk: true,
		},
		{
			name: "falls back to first alphabetical",
			props: notionapi.PropertyConfigs{
				"Zeta":  notionapi.DatePropertyConfig{Type: notionapi.PropertyConfigTypeDate},
				"Alpha": notionapi.DatePropertyConfig{Type: notionapi.PropertyConfigTypeDate},
			},
			want: "Alpha",
			weOk: true,
		},
		{
			name: "no date property",
			props: notionapi.PropertyConfigs{
				"Name": notionapi.TitlePropertyConfig{Type: notionapi.PropertyConfigTypeTitle},
			},
			want: "",
			weOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveDateProperty(&notionapi.Database{Properties: tt.props})
			if ok != tt.weOk || got != tt.want {
				t.Errorf("resolveDateProperty() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.weOk)
			}
		})
	}
}

func TestDbNoteCommandExecute(t *testing.T) {
	fixedNow := time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		message   string
		databases map[string]*notionapi.Database
		mockErr   error
		want      []string
		// assertions on the created page (nil = don't create / don't check)
		wantTitle   string
		wantDateSet bool
		wantDate    string
		wantStatus  string
		wantSelect  string
	}{
		{
			name:        "creates row with title only",
			message:     `/dbnote "Finance" coffee 3.50`,
			databases:   map[string]*notionapi.Database{"finance": financeDB()},
			want:        []string{"entry added to Finance"},
			wantTitle:   "coffee 3.50",
			wantDateSet: false,
		},
		{
			name:        "today stamps the date property",
			message:     `/dbnote "Finance" coffee --today`,
			databases:   map[string]*notionapi.Database{"finance": financeDB()},
			want:        []string{"entry added to Finance"},
			wantTitle:   "coffee",
			wantDateSet: true,
			wantDate:    "2026-08-29",
		},
		{
			name:        "explicit date and properties",
			message:     `/dbnote "Finance" lunch --date "2026-01-02" --status "Paid" --select "Food"`,
			databases:   map[string]*notionapi.Database{"finance": financeDB()},
			want:        []string{"entry added to Finance"},
			wantTitle:   "lunch",
			wantDateSet: true,
			wantDate:    "2026-01-02",
			wantStatus:  "Paid",
			wantSelect:  "Food",
		},
		{
			name:      "database not found",
			message:   `/dbnote "Groceries" milk`,
			databases: map[string]*notionapi.Database{},
			want:      []string{boterrors.ErrDatabaseNotFound.Error()},
		},
		{
			name:    "picks the named database, not the first fuzzy result",
			message: `/dbnote "Expenses" coffee 3.50`,
			databases: map[string]*notionapi.Database{
				"postits":  namedDB("db-post", "Post-its"),
				"expenses": namedDB("db-exp", "Expenses"),
			},
			want:      []string{"entry added to Expenses"},
			wantTitle: "coffee 3.50",
		},
		{
			name:    "missing quotes",
			message: `/dbnote Finance milk`,
			want:    []string{`Make sure you have enclosed the database name in quotes, e.g. /dbnote "Finance" coffee 3.50`},
		},
		{
			name:    "no title",
			message: `/dbnote "Finance"`,
			want:    []string{`Please provide the text for the new entry, e.g. /dbnote "Finance" coffee 3.50`},
		},
		{
			name:    "bad date format",
			message: `/dbnote "Finance" lunch --date "29-08-2026"`,
			want:    []string{`The date must be in YYYY-MM-DD format, e.g. --date "2026-08-29"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := notion.NewNotionMockWithDatabases(nil, tt.databases, tt.mockErr)
			b := bot(withUserRepo(db.NewUserRepoMock(map[int]*ent.User{1: {ID: 1}}, nil)))

			cmd := &DbNoteCommand{
				IBot: b,
				buildNotionClient: func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error) {
					return mock, nil
				},
				now: func() time.Time { return fixedNow },
			}
			cmd.Execute(context.Background(), update(withMessage(tt.message), withId(1)))

			if len(b.Resp) != len(tt.want) {
				t.Fatalf("responses = %v, want %v", b.Resp, tt.want)
			}
			for i := range tt.want {
				if b.Resp[i] != tt.want[i] {
					t.Errorf("response[%d] = %q, want %q", i, b.Resp[i], tt.want[i])
				}
			}

			if tt.wantTitle == "" {
				return
			}
			created := mock.CreatedPage()
			if created == nil {
				t.Fatal("expected a page to be created")
			}
			if got := propTitle(created.Properties["Name"]); got != tt.wantTitle {
				t.Errorf("title = %q, want %q", got, tt.wantTitle)
			}
			dateProp, hasDate := created.Properties["Date"].(notionapi.DateProperty)
			if hasDate != tt.wantDateSet {
				t.Errorf("date set = %v, want %v", hasDate, tt.wantDateSet)
			}
			if tt.wantDateSet {
				got := time.Time(*dateProp.Date.Start).Format(dateLayout)
				if got != tt.wantDate {
					t.Errorf("date = %q, want %q", got, tt.wantDate)
				}
			}
			if tt.wantStatus != "" {
				sp, _ := created.Properties["Status"].(notionapi.StatusProperty)
				if sp.Status.Name != tt.wantStatus {
					t.Errorf("status = %q, want %q", sp.Status.Name, tt.wantStatus)
				}
			}
			if tt.wantSelect != "" {
				sp, _ := created.Properties["Tag"].(notionapi.SelectProperty)
				if sp.Select.Name != tt.wantSelect {
					t.Errorf("select = %q, want %q", sp.Select.Name, tt.wantSelect)
				}
			}
		})
	}
}

func propTitle(p notionapi.Property) string {
	tp, ok := p.(notionapi.TitleProperty)
	if !ok || len(tp.Title) == 0 || tp.Title[0].Text == nil {
		return ""
	}
	return tp.Title[0].Text.Content
}
