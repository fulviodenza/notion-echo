package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jomei/notionapi"
	"github.com/mymmrac/telego"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	notionerrors "github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*DbNoteCommand)(nil)

const dateLayout = "2006-01-02"

// dashes are the accepted spellings of a flag prefix. Telegram and some
// keyboards turn "--" into an em/en dash, so we accept all of them.
var dashes = []string{"--", "—", "–"}

// statusPropertyConfigType is the config type Notion reports for a status
// property. notionapi has no exported constant for it.
const statusPropertyConfigType notionapi.PropertyConfigType = "status"

type DbNoteCommand struct {
	types.IBot
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)
	now               func() time.Time
	// uploadAttachment downloads the message's photo/document from Telegram and
	// uploads it to Notion, returning the file upload ID. Swappable in tests.
	uploadAttachment func(ctx context.Context, msg *telego.Message, userID int) (string, error)
}

func NewDbNoteCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)) types.Command {
	hc := DbNoteCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
		now:               time.Now,
	}
	hc.uploadAttachment = hc.downloadAndUploadAttachment
	return hc.Execute
}

type dbNoteRequest struct {
	dbName string
	title  string
	status string
	sel    string
	date   string
	today  bool
}

func (dc *DbNoteCommand) Execute(ctx context.Context, update *telego.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
	dc.Logger().Infof("[DbNoteCommand] got dbnote request from %d", id)

	metrics.DbNoteCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	// photos and documents carry the command in their caption
	messageText := update.Message.Text
	if messageText == "" {
		messageText = update.Message.Caption
	}
	hasAttachment := messageHasAttachment(update.Message)

	req, err := parseDbNoteMessage(messageText)
	if err != nil {
		dc.SendMessage("Make sure you have enclosed the database name in quotes, e.g. /dbnote \"Finance\" coffee 3.50", id, false, true)
		return
	}
	if req.dbName == "" {
		dc.SendMessage("No database name specified.", id, false, true)
		return
	}
	if req.title == "" && hasAttachment {
		req.title = attachmentTitle(update.Message)
	}
	if req.title == "" {
		dc.SendMessage("Please provide the text for the new entry, e.g. /dbnote \"Finance\" coffee 3.50", id, false, true)
		return
	}
	if req.date != "" {
		if _, perr := time.Parse(dateLayout, req.date); perr != nil {
			dc.SendMessage("The date must be in YYYY-MM-DD format, e.g. --date \"2026-08-29\"", id, false, true)
			return
		}
	}

	notionToken, err := dc.GetUserRepo().GetNotionTokenByID(ctx, id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("dbnote error")
		dc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}
	notionClient, err := dc.buildNotionClient(ctx, dc.GetUserRepo(), id, notionToken)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("dbnote error")
		dc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}

	databases, err := notionClient.SearchDatabase(ctx, req.dbName)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("dbnote error")
		dc.SendMessage(notionerrors.ErrDatabaseNotFound.Error(), id, false, true)
		return
	}
	if len(databases) == 0 {
		dc.SendMessage(notionerrors.ErrDatabaseNotFound.Error(), id, false, true)
		return
	}
	database, ok := selectDatabase(databases, req.dbName)
	if !ok {
		dc.SendMessage(notionerrors.ErrDatabaseNotFound.Error(), id, false, true)
		return
	}

	props, warnings, err := dc.buildProperties(database, req)
	if err != nil {
		dc.SendMessage(err.Error(), id, false, true)
		return
	}

	var children []notionapi.Block
	if hasAttachment {
		dc.SendMessage("uploading your file to Notion...", id, false, true)
		fileID, uerr := dc.uploadAttachment(ctx, update.Message, id)
		if uerr != nil {
			dc.Logger().WithFields(logrus.Fields{"error": uerr}).Error("dbnote error")
			dc.SendMessage("file error", id, false, true)
			return
		}
		children = attachFile(database, props, fileID, attachmentTitle(update.Message), update.Message)
	}

	_, err = notionClient.CreatePage(ctx, &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: notionapi.DatabaseID(database.ID),
		},
		Properties: props,
		Children:   children,
	})
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("dbnote error")
		dc.SendMessage(notionerrors.ErrSaveEntry.Error(), id, false, true)
		return
	}

	msg := fmt.Sprintf("entry added to %s", notion.ExtractRichText(database.Title))
	if len(warnings) > 0 {
		msg += " (" + strings.Join(warnings, ", ") + ")"
	}
	dc.SendMessage(msg, id, false, false)
}

// buildProperties maps the parsed request onto the database schema. It always
// sets the title property; status/select/date are only set when both the flag
// was provided and the schema has a matching property (otherwise a warning is
// collected). --today and --date resolve to the same date property.
func (dc *DbNoteCommand) buildProperties(database *notionapi.Database, req dbNoteRequest) (notionapi.Properties, []string, error) {
	props := notionapi.Properties{}
	warnings := []string{}

	titleName, ok := findPropByType(database, notionapi.PropertyConfigTypeTitle)
	if !ok {
		return nil, nil, notionerrors.ErrSaveEntry
	}
	props[titleName] = notionapi.TitleProperty{
		Title: []notionapi.RichText{{Text: &notionapi.Text{Content: req.title}}},
	}

	if req.status != "" {
		if name, ok := findPropByType(database, statusPropertyConfigType); ok {
			props[name] = notionapi.StatusProperty{Status: notionapi.Status{Name: req.status}}
		} else {
			warnings = append(warnings, "no status property, skipped --status")
		}
	}

	if req.sel != "" {
		if name, ok := findPropByType(database, notionapi.PropertyConfigTypeSelect); ok {
			props[name] = notionapi.SelectProperty{Select: notionapi.Option{Name: req.sel}}
		} else {
			warnings = append(warnings, "no select property, skipped --select")
		}
	}

	if req.today || req.date != "" {
		if name, ok := resolveDateProperty(database); ok {
			t := dc.now()
			if req.date != "" {
				t, _ = time.Parse(dateLayout, req.date) // already validated in Execute
			}
			d := notionapi.Date(t)
			props[name] = notionapi.DateProperty{Date: &notionapi.DateObject{Start: &d}}
		} else {
			warnings = append(warnings, "no date property, skipped the date")
		}
	}

	return props, warnings, nil
}

func messageHasAttachment(msg *telego.Message) bool {
	return (msg.Document != nil && msg.Document.FileID != "") || len(msg.Photo) > 0
}

// attachmentTitle is the name shown for the uploaded file, also used as the
// entry title when the caption only names the database.
func attachmentTitle(msg *telego.Message) string {
	if msg.Document != nil && msg.Document.FileName != "" {
		return msg.Document.FileName
	}
	if msg.Document != nil {
		return "file"
	}
	return "photo"
}

func (dc *DbNoteCommand) downloadAndUploadAttachment(ctx context.Context, msg *telego.Message, userID int) (string, error) {
	if msg.Document != nil && msg.Document.FileID != "" {
		return downloadAndUploadDocument(ctx, dc.IBot, msg.Document, dc.buildNotionClient, dc.GetUserRepo(), userID)
	}
	return downloadAndUploadImage(ctx, dc.IBot, getLargestPhotoSize(msg.Photo), dc.buildNotionClient, dc.GetUserRepo(), userID)
}

// attachFile puts the uploaded file in the database's "Files & media" property
// when it has one, so it shows up in table/gallery views. Otherwise it falls
// back to returning a block for the entry's page body.
func attachFile(database *notionapi.Database, props notionapi.Properties, fileID, name string, msg *telego.Message) []notionapi.Block {
	if propName, ok := findPropByType(database, notionapi.PropertyConfigTypeFiles); ok {
		props[propName] = FilesPropertyWithFileUpload{
			Type: notionapi.PropertyTypeFiles,
			Files: []FileWithFileUpload{{
				Name:       name,
				Type:       "file_upload",
				FileUpload: &FileUploadObject{ID: fileID},
			}},
		}
		return nil
	}
	return []notionapi.Block{buildBlockWithUploadedFile(fileID, msg.Document, msg.Photo)}
}

// FilesPropertyWithFileUpload is a files property value referencing a Notion
// file upload, which notionapi.FilesProperty can't express.
type FilesPropertyWithFileUpload struct {
	Type  notionapi.PropertyType `json:"type,omitempty"`
	Files []FileWithFileUpload   `json:"files"`
}

func (p FilesPropertyWithFileUpload) GetID() string                   { return "" }
func (p FilesPropertyWithFileUpload) GetType() notionapi.PropertyType { return p.Type }

type FileWithFileUpload struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	FileUpload *FileUploadObject `json:"file_upload"`
}

// selectDatabase picks the database the user actually named. Notion's search is
// fuzzy and ranks by relevance, so databases[0] is often the wrong one when the
// account has several similarly named databases (e.g. asking for "Expenses" but
// "Post-its" ranks first). We therefore prefer an exact, case-insensitive title
// match and only fall back to the sole result when the search was unambiguous.
func selectDatabase(databases []*notionapi.Database, name string) (*notionapi.Database, bool) {
	for _, db := range databases {
		if strings.EqualFold(strings.TrimSpace(notion.ExtractRichText(db.Title)), strings.TrimSpace(name)) {
			return db, true
		}
	}
	if len(databases) == 1 {
		return databases[0], true
	}
	return nil, false
}

// parseDbNoteMessage extracts the target database name, the entry title, and any
// property flags from a /dbnote message. The database name must be the first
// quoted token; flags may appear anywhere.
func parseDbNoteMessage(messageText string) (dbNoteRequest, error) {
	text := strings.TrimSpace(strings.Replace(messageText, "/dbnote", "", 1))

	var req dbNoteRequest
	text, req.today = extractBoolFlag(text, "today")
	req.status, text, _ = extractQuotedFlag(text, "status")
	req.sel, text, _ = extractQuotedFlag(text, "select")
	req.date, text, _ = extractQuotedFlag(text, "date")

	parts := strings.SplitN(text, "\"", 3)
	if len(parts) < 3 {
		return dbNoteRequest{}, notionerrors.ErrNotEnoughArguments
	}
	req.dbName = strings.TrimSpace(parts[1])
	req.title = strings.TrimSpace(parts[2])
	return req, nil
}

// extractQuotedFlag finds --name "value" (any dash spelling) anywhere in text,
// returns the value and text with that segment removed.
func extractQuotedFlag(text, name string) (value, rest string, found bool) {
	for _, dash := range dashes {
		flag := dash + name
		idx := strings.Index(text, flag)
		if idx == -1 {
			continue
		}
		parts := strings.SplitN(text[idx+len(flag):], "\"", 3)
		if len(parts) < 3 {
			continue
		}
		value = strings.TrimSpace(parts[1])
		rest = strings.TrimSpace(strings.TrimSpace(text[:idx]) + " " + strings.TrimSpace(parts[2]))
		return value, rest, true
	}
	return "", text, false
}

// extractBoolFlag removes a valueless --name flag (any dash spelling) from text.
func extractBoolFlag(text, name string) (string, bool) {
	for _, dash := range dashes {
		flag := dash + name
		if idx := strings.Index(text, flag); idx != -1 {
			return strings.TrimSpace(strings.TrimSpace(text[:idx]) + " " + strings.TrimSpace(text[idx+len(flag):])), true
		}
	}
	return text, false
}

// findPropByType returns the name of the first (alphabetically) property of the
// given config type in the database schema.
func findPropByType(database *notionapi.Database, t notionapi.PropertyConfigType) (string, bool) {
	names := []string{}
	for name, cfg := range database.Properties {
		if cfg.GetType() == t {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", false
	}
	sort.Strings(names)
	return names[0], true
}

// resolveDateProperty picks which date property to stamp: a property named
// Date/Created/Day (case-insensitive) is preferred, otherwise the first date
// property in the schema.
func resolveDateProperty(database *notionapi.Database) (string, bool) {
	dateProps := []string{}
	for name, cfg := range database.Properties {
		if cfg.GetType() == notionapi.PropertyConfigTypeDate {
			dateProps = append(dateProps, name)
		}
	}
	if len(dateProps) == 0 {
		return "", false
	}
	sort.Strings(dateProps)
	for _, preferred := range []string{"date", "created", "day"} {
		for _, name := range dateProps {
			if strings.EqualFold(name, preferred) {
				return name, true
			}
		}
	}
	return dateProps[0], true
}
