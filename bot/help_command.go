package bot

import (
	"context"
	"log"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
)

var _ types.ICommand = (*HelpCommand)(nil)

type HelpCommand struct {
	types.IBot
}

func NewHelpCommand(bot types.IBot) types.Command {
	hc := HelpCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (hc *HelpCommand) Execute(ctx context.Context, update *objects.Update) {
	if hc == nil || hc.IBot == nil {
		return
	}
	helpMessage := hc.GetHelpMessage()

	err := hc.SendMessage(helpMessage, update, true)
	if err != nil {
		log.Printf("Failed to send help message: %v", err)
	}
}
