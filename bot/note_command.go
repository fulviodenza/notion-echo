package bot

import (
	"context"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
)

var _ types.ICommand = (*NoteCommand)(nil)

const (
	NOTE_SAVED    = "note saved!"
	NO_VALID_TAGS = "no valid tags found"
)

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
	// args := strings.Split(update.Message.Text, " ")
	cc.Bot.SendMessage(NOTE_SAVED, update, false)
}
