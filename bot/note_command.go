package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
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
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)
}

func NewNoteCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)) types.Command {
	hc := NoteCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
	}
	return hc.Execute
}

func (cc *NoteCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if cc == nil || cc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
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
		if noteText == "" && update.Message.Document == nil && update.Message.Photo == nil {
			cc.SetUserState(id, "/note")
			cc.SendMessage("write your note in the next message", id, false, true)
			return
		}
	}
	// the noteText contains --page string
	if noteText == "" && update.Message.Document == nil && update.Message.Photo == nil {
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
	uploadedFileID := ""
	children := []notionapi.Block{}
	if update.Message.Document != nil && update.Message.Document.FileID != "" {
		uploadedFileID, err = downloadAndUploadDocument(ctx, cc.IBot, update.Message.Document, cc.buildNotionClient, cc.GetUserRepo(), id)
	}
	if update.Message.Photo != nil {
		cc.SendMessage(`uploading your photo to Notion...`, id, false, true)
		// Select the largest photo size instead of the first one
		largestPhoto := getLargestPhotoSize(update.Message.Photo)
		uploadedFileID, err = downloadAndUploadImage(ctx, cc.IBot, largestPhoto, cc.buildNotionClient, cc.GetUserRepo(), id)
	}
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage("file error", id, false, true)
		return
	}
	if uploadedFileID != "" {
		children = append(children, buildBlockWithUploadedFile(uploadedFileID, update.Message.Document, update.Message.Photo))
	}

	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText, children))

	notionToken, err := cc.GetUserRepo().GetNotionTokenByID(ctx, id)
	if err != nil {
		cc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		cc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}
	notionClient, err := cc.buildNotionClient(ctx, cc.GetUserRepo(), id, notionToken)
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

func downloadAndUploadDocument(ctx context.Context, bot types.IBot, ps *tgbotapi.Document, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error), userRepo db.UserRepoInterface, userID int) (string, error) {
	file, err := bot.GetTelegramClient().GetFile(tgbotapi.FileConfig{
		FileID: ps.FileID,
	})
	if err != nil {
		return "", err
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("TELEGRAM_TOKEN"), file.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	notionToken, err := userRepo.GetNotionTokenByID(ctx, userID)
	if err != nil {
		return "", err
	}

	notionClient, err := buildNotionClient(ctx, userRepo, userID, notionToken)
	if err != nil {
		return "", err
	}

	fileName := ps.FileName
	if fileName == "" {
		fileName = ps.FileID + filepath.Ext(file.FilePath)
	}

	uploadResp, err := notionClient.UploadFile(ctx, fileName, fileData)
	if err != nil {
		return "", err
	}

	return uploadResp.ID, nil
}

func downloadAndUploadImage(ctx context.Context, bot types.IBot, ps tgbotapi.PhotoSize, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error), userRepo db.UserRepoInterface, userID int) (string, error) {
	file, err := bot.GetTelegramClient().GetFile(tgbotapi.FileConfig{
		FileID: ps.FileID,
	})
	if err != nil {
		return "", err
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("TELEGRAM_TOKEN"), file.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	notionToken, err := userRepo.GetNotionTokenByID(ctx, userID)
	if err != nil {
		return "", err
	}

	notionClient, err := buildNotionClient(ctx, userRepo, userID, notionToken)
	if err != nil {
		return "", err
	}

	fileName := ps.FileID + filepath.Ext(file.FilePath)
	uploadResp, err := notionClient.UploadFile(ctx, fileName, fileData)
	if err != nil {
		return "", err
	}

	return uploadResp.ID, nil
}

func buildBlockWithUploadedFile(fileID string, document *tgbotapi.Document, photo []tgbotapi.PhotoSize) (b notionapi.Block) {
	if document != nil {
		ext := utils.GetExt(document.FileName)
		switch ext {
		case "pdf":
			b = buildPdfBlockWithFileUpload(fileID)
		default:
			b = buildFileBlockWithFileUpload(fileID)
		}
	} else if photo != nil {
		b = buildImageBlockWithFileUpload(fileID)
	}
	return b
}

func buildPdfBlockWithFileUpload(fileID string) *PdfBlockWithFileUpload {
	file := &PdfBlockWithFileUpload{
		BasicBlock: notionapi.BasicBlock{
			Object: "block",
			Type:   notionapi.BlockTypePdf,
		},
		Pdf: PdfWithFileUpload{
			Type: "file_upload",
			FileUpload: &FileUploadObject{
				ID: fileID,
			},
		},
	}
	return file
}

func buildFileBlockWithFileUpload(fileID string) *FileBlockWithFileUpload {
	file := &FileBlockWithFileUpload{
		BasicBlock: notionapi.BasicBlock{
			Object: "block",
			Type:   notionapi.BlockTypeFile,
		},
		File: BlockFileWithFileUpload{
			Type: "file_upload",
			FileUpload: &FileUploadObject{
				ID: fileID,
			},
		},
	}
	return file
}

func buildImageBlockWithFileUpload(fileID string) *ImageBlockWithFileUpload {
	image := &ImageBlockWithFileUpload{
		BasicBlock: notionapi.BasicBlock{
			Object: "block",
			Type:   notionapi.BlockTypeImage,
		},
		Image: ImageWithFileUpload{
			Type: "file_upload",
			FileUpload: &FileUploadObject{
				ID: fileID,
			},
		},
	}
	return image
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

func getLargestPhotoSize(photoSizes []tgbotapi.PhotoSize) tgbotapi.PhotoSize {
	var largestPhoto tgbotapi.PhotoSize
	var maxSize int

	for _, size := range photoSizes {
		if size.FileSize > maxSize {
			maxSize = size.FileSize
			largestPhoto = size
		}
	}

	return largestPhoto
}

// Custom types to support file uploads since notionapi doesn't have them yet
type FileUploadObject struct {
	ID string `json:"id"`
}

type PdfWithFileUpload struct {
	Type       string                `json:"type"`
	FileUpload *FileUploadObject     `json:"file_upload,omitempty"`
	File       *notionapi.FileObject `json:"file,omitempty"`
}

type BlockFileWithFileUpload struct {
	Type       string                `json:"type"`
	FileUpload *FileUploadObject     `json:"file_upload,omitempty"`
	File       *notionapi.FileObject `json:"file,omitempty"`
}

type ImageWithFileUpload struct {
	Type       string                `json:"type"`
	FileUpload *FileUploadObject     `json:"file_upload,omitempty"`
	File       *notionapi.FileObject `json:"file,omitempty"`
}

// Custom block types
type PdfBlockWithFileUpload struct {
	notionapi.BasicBlock
	Pdf PdfWithFileUpload `json:"pdf"`
}

type FileBlockWithFileUpload struct {
	notionapi.BasicBlock
	File BlockFileWithFileUpload `json:"file"`
}

type ImageBlockWithFileUpload struct {
	notionapi.BasicBlock
	Image ImageWithFileUpload `json:"image"`
}
