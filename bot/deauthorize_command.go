package bot

import (
	"context"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
)

var _ types.ICommand = (*DeauthorizeCommand)(nil)

type DeauthorizeCommand struct {
	types.IBot
}

func NewDeauthorizeCommand(bot *Bot) types.Command {
	hc := DefaultPageCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (dc *DeauthorizeCommand) Execute(ctx context.Context, update *objects.Update) {
	err := dc.GetUserRepo().DeleteUser(ctx, update.Message.Chat.Id)
	if err != nil {
		dc.SendMessage("error deleting user", update, true)
		return
	}
	dc.SendMessage("deleted user!", update, true)
}
