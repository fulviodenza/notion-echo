package bot

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/notion-echo/utils"
)

var _ types.ICommand = (*NoteCommand)(nil)

const (
	NOTE_SAVED = "note saved!"
)

var BotEmoji = notionapi.Emoji("🤖")

type NoteCommand struct {
	types.IBot
}

func NewNoteCommand(bot *Bot) types.Command {
	hc := NoteCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (cc *NoteCommand) Execute(ctx context.Context, update *objects.Update) {
	if cc == nil || cc.IBot == nil {
		return
	}

	userRepo := cc.GetUserRepo()
	id := update.Message.Chat.Id

	blocks := &notionapi.AppendBlockChildrenRequest{}

	tokenEnc, err := userRepo.GetNotionTokenByID(ctx, id)
	if err != nil || tokenEnc == "" {
		return
	}
	encKey, err := cc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		log.Println(err)
		return
	}
	token, err := utils.DecryptString(tokenEnc, encKey)
	if err != nil {
		log.Println(err)
		return
	}

	defaultPage, err := userRepo.GetDefaultPage(ctx, id)
	if err != nil {
		log.Println(err)
		return
	}
	if defaultPage == "" {
		cc.SendMessage("first select an authorized page from your Notion!", update, false)
		return
	}
	notionClient := notion.NewNotionService(notionapi.NewClient(notionapi.Token(token)))
	page, err := notionClient.SearchPage(ctx, defaultPage)
	if err != nil {
		return
	}

	paths, err := downloadAndUploadImage(cc.IBot, update.Message.Photo)
	if err != nil {
		return
	}
	for _, fp := range paths {
		blocks.Children = append(blocks.Children, buildImageBlock(fp))
	}

	noteText := strings.Replace(update.Message.Text, "/note", "", 1)
	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText))

	_, err = notionClient.Block().AppendChildren(ctx, notionapi.BlockID(page.ID), blocks)
	if err != nil {
		log.Println(err)
		cc.SendMessage(errors.ErrSaveNote.Error(), update, false)
		return
	}

	cc.SendMessage(NOTE_SAVED, update, false)
}

func downloadAndUploadImage(bot types.IBot, ps []objects.PhotoSize) ([]string, error) {
	var filePaths []string = make([]string, len(ps))
	for _, p := range ps {
		out, err := os.Create(p.FileId)
		if err != nil {
			continue
		}
		file, err := bot.GetTelegramClient().GetFile(p.FileId, true, out)
		if err != nil {
			continue
		}
		filePaths = append(filePaths, file.FilePath)
	}
	return filePaths, nil
}

func buildImageBlock(path string) *notionapi.ImageBlock {
	image := &notionapi.ImageBlock{
		BasicBlock: notionapi.BasicBlock{
			Type:   notionapi.BlockTypeImage,
			Object: "block",
		},
		Image: notionapi.Image{
			Type: "external",
			File: &notionapi.FileObject{
				URL: path,
			},
		},
	}

	return image
}
func buildCalloutBlock(text string) *notionapi.CalloutBlock {
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
			Children: nil,
		},
	}

	return callout
}
