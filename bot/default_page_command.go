package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
)

var _ types.ICommand = (*DefaultPageCommand)(nil)

type DefaultPageCommand struct {
	types.IBot
}

func NewDefaultPageCommand(bot *Bot) types.Command {
	hc := DefaultPageCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (dc *DefaultPageCommand) Execute(ctx context.Context, update *objects.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}
	args := strings.Split(update.Message.Text, " ")
	if len(args) != 2 {
		dc.SendMessage(errors.ErrNotEnoughArguments.Error(), update, false)
		return
	}
	selectedPage := args[1]
	err := dc.GetUserRepo().SetDefaultPage(ctx, update.Message.Chat.Id, selectedPage)
	if err != nil {
		log.Println(err)
		dc.SendMessage(errors.ErrSetDefaultPage.Error(), update, false)
		return
	}
	dc.SendMessage(fmt.Sprintf("page %s set as default!", selectedPage), update, false)
}
