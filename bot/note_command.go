package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/notion-echo/utils"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*NoteCommand)(nil)

const (
	NOTE_SAVED = "note saved!"
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

	blocks := &notionapi.AppendBlockChildrenRequest{}

	messageText := update.Message.Text
	if update.Message.Caption != "" {
		messageText = update.Message.Caption
	}
	noteText := strings.Replace(messageText, "/note", "", 1)
	if noteText == "" && update.Message.Text != "" {
		cc.SendMessage("write something in your note!", id, false)
		return
	}

	children := []notionapi.Block{}
	filePath := ""
	var err error
	if update.Message.Document != nil && update.Message.Document.FileId != "" {
		filePath, err = downloadAndUploadDocument(cc.IBot, update.Message.Document)
	}
	if update.Message.Photo != nil {
		cc.SendMessage(`ensure your photo is sent without compression!, 
			I will save it for you but you could have issues in visualizing it`, id, false)
		filePath, err = downloadAndUploadImage(cc.IBot, update.Message.Photo[0])
	}
	if err != nil {
		cc.SendMessage("file error!", id, false)
		return
	}
	if filePath != "" {
		children = append(children, buildBlock(filePath))
	}

	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText, children))

	encKey, err := cc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(errors.ErrInternal.Error(), id, false)
		return
	}
	notionClient, err := cc.buildNotionClient(ctx, cc.GetUserRepo(), id, encKey)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(errors.ErrNotRegistered.Error(), id, false)
		return
	}
	defaultPage, err := cc.GetUserRepo().GetDefaultPage(ctx, id)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
	}
	if err != nil || defaultPage == "" {
		cc.SendMessage(errors.ErrPageNotFound.Error(), id, false)
		return
	}
	page, err := notionClient.SearchPage(ctx, defaultPage)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(errors.ErrPageNotFound.Error(), id, false)
		return
	}
	_, err = notionClient.Block().AppendChildren(ctx, notionapi.BlockID(page.ID), blocks)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(errors.ErrSaveNote.Error(), id, false)
		return
	}
	cc.SendMessage(NOTE_SAVED, id, false)
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
