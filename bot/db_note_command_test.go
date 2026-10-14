package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jomei/notionapi"
	"github.com/mymmrac/telego"
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

func TestDbNoteCommandAttachment(t *testing.T) {
	receiptsDB := func() *notionapi.Database {
		d := namedDB("db-rec", "Receipts")
		d.Properties["Scan"] = notionapi.FilesPropertyConfig{Type: notionapi.PropertyConfigTypeFiles}
		return d
	}
	photo := []telego.PhotoSize{{FileID: "small", FileSize: 10}, {FileID: "big", FileSize: 100}}

	tests := []struct {
		name      string
		caption   string
		database  *notionapi.Database
		document  *telego.Document
		photo     []telego.PhotoSize
		uploadErr error
		want      []string
		wantTitle string
		// wantProp: file lands in the "Scan" files property; otherwise in a body block
		wantProp      bool
		wantBlockType notionapi.BlockType
	}{
		{
			name:      "photo goes into the files property",
			caption:   `/dbnote "Receipts" coffee`,
			database:  receiptsDB(),
			photo:     photo,
			want:      []string{"uploading your file to Notion...", "entry added to Receipts"},
			wantTitle: "coffee",
			wantProp:  true,
		},
		{
			name:          "photo falls back to an image block without a files property",
			caption:       `/dbnote "Expenses" coffee`,
			database:      namedDB("db-exp", "Expenses"),
			photo:         photo,
			want:          []string{"uploading your file to Notion...", "entry added to Expenses"},
			wantTitle:     "coffee",
			wantBlockType: notionapi.BlockTypeImage,
		},
		{
			name:          "pdf falls back to a pdf block and titles the entry after the file",
			caption:       `/dbnote "Expenses"`,
			database:      namedDB("db-exp", "Expenses"),
			document:      &telego.Document{FileID: "doc", FileName: "invoice.pdf"},
			want:          []string{"uploading your file to Notion...", "entry added to Expenses"},
			wantTitle:     "invoice.pdf",
			wantBlockType: notionapi.BlockTypePdf,
		},
		{
			name:      "upload failure creates no entry",
			caption:   `/dbnote "Receipts" coffee`,
			database:  receiptsDB(),
			photo:     photo,
			uploadErr: errors.New("telegram down"),
			want:      []string{"uploading your file to Notion...", "file error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := notion.NewNotionMockWithDatabases(nil, map[string]*notionapi.Database{"db": tt.database}, nil)
			b := bot(withUserRepo(db.NewUserRepoMock(map[int]*ent.User{1: {ID: 1}}, nil)))

			var uploaded *telego.Message
			cmd := &DbNoteCommand{
				IBot: b,
				buildNotionClient: func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error) {
					return mock, nil
				},
				now: time.Now,
				uploadAttachment: func(ctx context.Context, msg *telego.Message, userID int) (string, error) {
					uploaded = msg
					return "upload-1", tt.uploadErr
				},
			}
			up := update(withId(1))
			up.Message.Caption = tt.caption
			up.Message.Document = tt.document
			up.Message.Photo = tt.photo
			cmd.Execute(context.Background(), up)

			if len(b.Resp) != len(tt.want) {
				t.Fatalf("responses = %v, want %v", b.Resp, tt.want)
			}
			for i := range tt.want {
				if b.Resp[i] != tt.want[i] {
					t.Errorf("response[%d] = %q, want %q", i, b.Resp[i], tt.want[i])
				}
			}
			if uploaded == nil {
				t.Fatal("expected the attachment to be uploaded")
			}

			created := mock.CreatedPage()
			if tt.wantTitle == "" {
				if created != nil {
					t.Fatal("expected no page to be created")
				}
				return
			}
			if created == nil {
				t.Fatal("expected a page to be created")
			}
			if got := propTitle(created.Properties["Name"]); got != tt.wantTitle {
				t.Errorf("title = %q, want %q", got, tt.wantTitle)
			}

			fp, hasProp := created.Properties["Scan"].(FilesPropertyWithFileUpload)
			if hasProp != tt.wantProp {
				t.Fatalf("files property set = %v, want %v", hasProp, tt.wantProp)
			}
			if tt.wantProp {
				if len(fp.Files) != 1 || fp.Files[0].FileUpload.ID != "upload-1" || fp.Files[0].Type != "file_upload" {
					t.Errorf("files property = %+v", fp)
				}
				if len(created.Children) != 0 {
					t.Errorf("expected no body blocks, got %d", len(created.Children))
				}
				return
			}
			if len(created.Children) != 1 || created.Children[0].GetType() != tt.wantBlockType {
				t.Fatalf("children = %+v, want one %s block", created.Children, tt.wantBlockType)
			}
		})
	}
}
