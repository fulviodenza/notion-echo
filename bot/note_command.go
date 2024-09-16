package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	notionerrors "github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/notion-echo/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*NoteCommand)(nil)

const (
	NOTE_SAVED = "note saved"
)

var BotEmoji = notionapi.Emoji("🤖")

type NoteCommand struct {
	types.IBot
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error)
}

func NewNoteCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error)) types.Command {
	hc := NoteCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
	}
	return hc.Execute
}

func (cc *NoteCommand) Execute(ctx context.Context, update *objects.Update) {
	if cc == nil || cc.IBot == nil {
		return
	}

	id := update.Message.Chat.Id
	cc.Logger().Infof("[NoteCommand] got note from %d", id)

	metrics.NoteCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	blocks := &notionapi.AppendBlockChildrenRequest{}

	messageText := update.Message.Text
	if update.Message.Caption != "" {
		messageText = update.Message.Caption
	}
	var pageName string
	var noteText string
	if !strings.Contains(messageText, "—page") {
		noteText = strings.Replace(messageText, "/note", "", 1)
		if noteText == "" {
			cc.SetUserState(id, "/note")
			cc.SendMessage("write your note in the next message", id, false, true)
			return
		}
	}
	// the noteText contains --page string
	if noteText == "" {
		parts := strings.SplitN(messageText, "\"", 3)
		if len(parts) < 3 {
			cc.SendMessage("Make sure you have enclosed the page name in quotes.", id, false, true)
			return
		}
		pageName = parts[1]
		noteText = parts[2]
		if pageName == "" {
			cc.SendMessage("No page name specified.", id, false, true)
			return
		}
		if noteText == "" {
			cc.SetUserState(id, "/note")
			cc.SendMessage("write your note in the next message", id, false, true)
			return
		}

	}
	defer func(userID int) {
		if cc.GetUserState(userID) != "" {
			cc.DeleteUserState(userID)
		}
	}(id)

	var err error
	filePath := ""
	children := []notionapi.Block{}
	if update.Message.Document != nil && update.Message.Document.FileId != "" {
		filePath, err = downloadAndUploadDocument(cc.IBot, update.Message.Document)
	}
	if update.Message.Photo != nil {
		cc.SendMessage(`ensure your photo is sent without compression!, 
			I will save it for you but you could have issues in visualizing it`, id, false, true)
		filePath, err = downloadAndUploadImage(cc.IBot, update.Message.Photo[0])
	}
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage("file error", id, false, true)
		return
	}
	if filePath != "" {
		children = append(children, buildBlock(filePath))
	}

	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText, children))

	encKey, err := cc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrInternal.Error(), id, false, true)
		return
	}
	notionClient, err := cc.buildNotionClient(ctx, cc.GetUserRepo(), id, encKey)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}

	if pageName == "" {
		defaultPage, err := cc.GetUserRepo().GetDefaultPage(ctx, id)
		// we ignore the err not found because if we cannot find the page in the db
		// the empty string will still look for all pages the bot has access to and select
		// the first one to write on
		if err != nil && !ent.IsNotFound(err) {
			cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
			cc.SendMessage(notionerrors.ErrPageNotFound.Error(), id, false, true)
			return
		}
		pageName = defaultPage
	}
	pages, err := notionClient.SearchPage(ctx, pageName)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrPageNotFound.Error(), id, false, true)
		return
	}
	if len(pages) == 0 {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrBotNotAuthorized.Error(), id, false, true)
		return
	}
	page := pages[0]

	_, err = notionClient.Block().AppendChildren(ctx, notionapi.BlockID(page.ID), blocks)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrSaveNote.Error(), id, false, true)
		return
	}

	cc.SendMessage(fmt.Sprintf("%s on %s page", NOTE_SAVED, notion.ExtractName(page.Properties)), id, false, false)
}

func downloadAndUploadDocument(bot types.IBot, ps *objects.Document) (string, error) {
	out, err := os.Create(ps.FileId)
	if err != nil {
		return "", err
	}

	// This get file is performed to be able to get
	// the filename to retrieve
	file, err := bot.GetTelegramClient().GetFile(ps.FileId, true, out)
	if err != nil {
		return "", err
	}
	return file.FilePath, os.Remove(ps.FileId)
}

func downloadAndUploadImage(bot types.IBot, ps objects.PhotoSize) (string, error) {
	out, err := os.Create(ps.FileId)
	if err != nil {
		return "", err
	}
	file, err := bot.GetTelegramClient().GetFile(ps.FileId, true, out)
	if err != nil {
		return "", err
	}
	return file.FilePath, nil
}

func buildBlock(path string) (b notionapi.Block) {
	ext := utils.GetExt(path)
	// bot-allowed file extensions
	switch ext {
	case "pdf":
		b = buildPdfBlock(path)
	case "png", "jpg", "jpeg":
		b = buildImageBlock(path)
	}
	return b
}

func buildCalloutBlock(text string, children []notionapi.Block) *notionapi.CalloutBlock {
	callout := &notionapi.CalloutBlock{
		BasicBlock: notionapi.BasicBlock{
			Type:   notionapi.BlockCallout,
			Object: "block",
		},
		Callout: notionapi.Callout{
			Icon: &notionapi.Icon{
				Type:  "emoji",
				Emoji: &BotEmoji,
			},
			RichText: []notionapi.RichText{
				{
					Text: &notionapi.Text{Content: text},
				},
			},
			Children: children,
		},
	}
	return callout
}

func buildPdfBlock(path string) *notionapi.PdfBlock {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("TELEGRAM_TOKEN"), path)
	file := &notionapi.PdfBlock{
		BasicBlock: notionapi.BasicBlock{
			Object: "block",
			Type:   notionapi.BlockTypePdf,
		},
		Pdf: notionapi.Pdf{
			Type: "external",
			External: &notionapi.FileObject{
				URL: url,
			},
		},
	}
	return file
}

func buildImageBlock(path string) *notionapi.ImageBlock {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("TELEGRAM_TOKEN"), path)
	image := &notionapi.ImageBlock{
		BasicBlock: notionapi.BasicBlock{
			Object: "block",
			Type:   notionapi.BlockTypeImage,
		},
		Image: notionapi.Image{
			Type: "external",
			External: &notionapi.FileObject{
				URL: url,
			},
		},
	}
	return image
}
