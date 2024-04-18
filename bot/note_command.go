package bot

import (
	"context"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
<<<<<<< Updated upstream
=======
	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
>>>>>>> Stashed changes
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
	noteText := strings.Replace(update.Message.Text, "/note", "", 1)
	if noteText == "" {
		cc.SendMessage("write something in your note!", update, false)
		return
	}
	blocks.Children = append(blocks.Children, buildCalloutBlock(noteText))

	encKey, err := cc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		return
	}
	notionClient, err := cc.buildNotionClient(ctx, cc.GetUserRepo(), id, encKey)
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
	blocks := markdownToNotionBlocks(text) // Convert Markdown to Notion blocks
	// Extract the rich text for non-list blocks and prepare list blocks to be added as children.
	var richTexts []notionapi.RichText
	var listBlocks []notionapi.Block
	for _, block := range blocks {
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			richTexts = blocksToRichTexts(blocks)
			richTexts = append(richTexts, b.Paragraph.RichText...)
		case *notionapi.BulletedListItemBlock, *notionapi.NumberedListItemBlock:
			// If we can include list blocks, append them directly to the children of the callout
			richTexts = blocksToRichTexts(blocks)
			listBlocks = append(listBlocks, block)
		}
	}

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
			RichText: richTexts,  // Set the rich text from non-list blocks
			Children: listBlocks, // Add list blocks directly if they are allowed as children
		},
	}
	return callout
}

func markdownToNotionBlocks(md string) []notionapi.Block {
	var blocks []notionapi.Block
	var currentBlock notionapi.Block
	var insideListItem bool // To track if we are currently inside a list item

	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)
	reader := text.NewReader([]byte(md))
	document := markdown.Parser().Parse(reader)
	ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch v := n.(type) {
		case *ast.Emphasis:
			if !entering {
				// Apply formatting styles after processing all child nodes
				applyFormatting(currentBlock, v.Level, false)
			}
		case *ast.ListItem:
			if entering {
				// Starting a new list item
				insideListItem = true
				currentBlock = createListItemBlock(v) // Use your function to create a list item block
			} else {
				// Exiting a list item
				blocks = append(blocks, currentBlock)
				insideListItem = false
				currentBlock = nil
			}
		case *ast.Text:
			textContent := string(v.Text(reader.Source()))
			trimmedContent := strings.TrimSpace(textContent)
			if trimmedContent != "" {
				if insideListItem && currentBlock != nil {
					// We're inside a list item, add text to the current list item block
					addToRichText(currentBlock, trimmedContent, v.PreviousSibling() == nil, v.NextSibling() == nil)
					insideListItem = false
				}
				if currentBlock == nil && !insideListItem {
					addToRichText(nil, trimmedContent, true, true)
				}
			}
		}
		return ast.WalkContinue, nil
	})

	return blocks
}

func blocksToRichTexts(blocks []notionapi.Block) []notionapi.RichText {
	var richTexts []notionapi.RichText
	for _, block := range blocks {
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			richTexts = append(richTexts, b.Paragraph.RichText...)
		case *notionapi.BulletedListItemBlock:
			// Add a bullet and a space before the list item content
			richTexts = append(richTexts, processListItem("- ")...)
		case *notionapi.NumberedListItemBlock:
			// For numbered lists, you would somehow need to keep track of the item number.
			// This is a more complex situation and may require additional state to track item numbers correctly.
			richTexts = append(richTexts, processListItem("1. ")...) // Simplified
		}
	}
	return richTexts
}

func processListItem(listItemPrefix string) []notionapi.RichText {
	var processedTexts []notionapi.RichText
	// Add the list item prefix (e.g., "- " for bullets, "1. " for numbers)
	processedTexts = append(processedTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: listItemPrefix,
		},
	})
	return processedTexts
}

func createListItemBlock(v *ast.ListItem) notionapi.Block {
	// Determine if the list is ordered or not
	isOrdered := v.Parent().(*ast.List).IsOrdered()
	richText := []notionapi.RichText{} // Placeholder for initial rich text content

	if isOrdered {
		return &notionapi.NumberedListItemBlock{
			BasicBlock: notionapi.BasicBlock{
				Type:   notionapi.BlockTypeNumberedListItem,
				Object: "block",
			},
			NumberedListItem: notionapi.ListItem{
				RichText: richText,
			},
		}
	} else {
		return &notionapi.BulletedListItemBlock{
			BasicBlock: notionapi.BasicBlock{
				Type:   notionapi.BlockTypeBulletedListItem,
				Object: "block",
			},
			BulletedListItem: notionapi.ListItem{
				RichText: richText,
			},
		}
	}
}

func addToRichText(block notionapi.Block, text string, isFirst bool, isLast bool) {
	var richTexts *[]notionapi.RichText
	if block == nil {
		richTexts = &[]notionapi.RichText{
			{
				Text: &notionapi.Text{
					Content: text,
				},
			},
		}
	}
	switch v := block.(type) {
	case *notionapi.ParagraphBlock:
		richTexts = &v.Paragraph.RichText
	case *notionapi.NumberedListItemBlock:
		richTexts = &v.NumberedListItem.RichText
	case *notionapi.BulletedListItemBlock:
		richTexts = &v.BulletedListItem.RichText
	}

	// Strip leading and trailing space if it's not the first or last segment
	if !isFirst {
		text = strings.TrimLeft(text, " \t")
	}
	if !isLast {
		text = strings.TrimRight(text, " \t")
	}

	*richTexts = append(*richTexts, notionapi.RichText{
		Text: &notionapi.Text{
			Content: text,
		},
	})
}

func applyFormatting(block notionapi.Block, level int, isItalic bool) {
	var richTexts *[]notionapi.RichText
	switch v := block.(type) {
	case *notionapi.ParagraphBlock:
		richTexts = &v.Paragraph.RichText
	case *notionapi.NumberedListItemBlock:
		richTexts = &v.NumberedListItem.RichText
	case *notionapi.BulletedListItemBlock:
		richTexts = &v.BulletedListItem.RichText
	}

	// Apply formatting to the last added rich text
	if len(*richTexts) > 0 {
		last := len(*richTexts) - 1
		if (*richTexts)[last].Annotations == nil {
			(*richTexts)[last].Annotations = &notionapi.Annotations{}
		}
		if isItalic {
			(*richTexts)[last].Annotations.Italic = true
		} else {
			if level == 2 {
				(*richTexts)[last].Annotations.Bold = true
			} // Add more conditions if needed
		}
	}
}
