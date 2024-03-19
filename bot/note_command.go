package bot

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/bot/types"
)

var _ types.ICommand = (*NoteCommand)(nil)

const (
	NOTE_SAVED    = "note saved!"
	SaveNoteErr   = "error saving note!"
	SearchPageErr = "writing page not found!"
)

var BotEmoji = notionapi.Emoji("🤖")

type NoteCommand struct {
	Bot types.IBot
}

func NewNoteCommand(bot *Bot) types.Command {
	hc := NoteCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (cc *NoteCommand) Execute(ctx context.Context, update *objects.Update) {
	blocks := &notionapi.AppendBlockChildrenRequest{}

	page, err := cc.Bot.GetNotionClient().SearchPage(ctx, "Buffer")
	if err != nil {
		return
	}

	paths, err := downloadAndUploadImage(cc.Bot, update.Message.Photo)
	if err != nil {
		return
	}
	for _, fp := range paths {
		blocks.Children = append(blocks.Children, buildImageBlock(fp))
	}

	noteText := strings.Replace(update.Message.Text, "/note", "", 1)
	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText))

	response, err := cc.Bot.GetNotionClient().Block().AppendChildren(ctx, notionapi.BlockID(page.ID), blocks)
	if err != nil {
		log.Println(err)
		cc.Bot.SendMessage(SaveNoteErr, update, false)
		return
	}

	log.Println("notion response: ", response)

	cc.Bot.SendMessage(NOTE_SAVED, update, false)
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
