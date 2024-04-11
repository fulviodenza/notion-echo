package bot

import (
	"context"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
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

	id := update.Message.Chat.Id

	blocks := &notionapi.AppendBlockChildrenRequest{}
	noteText := strings.Replace(update.Message.Text, "/note", "", 1)
	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText))

	encKey, err := cc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		return
	}
	notionClient, err := buildNotionClient(ctx, cc.GetUserRepo(), id, encKey)
	if err != nil {
		cc.SendMessage(errors.ErrNotRegistered.Error(), update, false)
		return
	}
	defaultPage, err := cc.GetUserRepo().GetDefaultPage(ctx, id)
	if err != nil {
		cc.SendMessage(errors.ErrPageNotFound.Error(), update, false)
		return
	}
	if defaultPage == "" {
		cc.SendMessage("first choose a default page between the authorized pages from your Notion!", update, false)
		return
	}
	page, err := notionClient.SearchPage(ctx, defaultPage)
	if err != nil {
		cc.SendMessage(errors.ErrPageNotFound.Error(), update, false)
		return
	}

	_, err = notionClient.Block().AppendChildren(ctx, notionapi.BlockID(page.ID), blocks)
	if err != nil {
		cc.SendMessage(errors.ErrSaveNote.Error(), update, false)
		return
	}

	cc.SendMessage(NOTE_SAVED, update, false)
}

// to enable note images we need to wait for the merge of the pr on SakoDroid telego
// func downloadAndUploadImage(bot types.IBot, ps []objects.PhotoSize) ([]string, error) {
// 	var filePaths []string = make([]string, len(ps))
// 	for _, p := range ps {
// 		out, err := os.Create(p.FileId)
// 		if err != nil {
// 			continue
// 		}
// 		file, err := bot.GetTelegramClient().GetFile(p.FileId, true, out)
// 		if err != nil {
// 			continue
// 		}
// 		filePaths = append(filePaths, file.FilePath)
// 	}
// 	return filePaths, nil
// }

// func buildImageBlock(path string) *notionapi.ImageBlock {
// 	image := &notionapi.ImageBlock{
// 		BasicBlock: notionapi.BasicBlock{
// 			Type:   notionapi.BlockTypeImage,
// 			Object: "block",
// 		},
// 		Image: notionapi.Image{
// 			Type: "external",
// 			File: &notionapi.FileObject{
// 				URL: path,
// 			},
// 		},
// 	}

// 	return image
// }

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
