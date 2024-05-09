package bot

import (
	"context"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/sirupsen/logrus"
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

	err := hc.SendMessage(helpMessage, update.Message.Chat.Id, true, true)
	if err != nil {
		hc.Logger().WithFields(logrus.Fields{"error": err}).Error("help error")
	}
}
