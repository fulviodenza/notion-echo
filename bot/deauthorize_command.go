package bot

import (
	"context"
	"fmt"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*DeauthorizeCommand)(nil)

type DeauthorizeCommand struct {
	types.IBot
}

func NewDeauthorizeCommand(bot types.IBot) types.Command {
	hc := DeauthorizeCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (dc *DeauthorizeCommand) Execute(ctx context.Context, update *objects.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := update.Message.Chat.Id
	dc.Logger().Infof("[DeauthorizeCommand] got deauthorize request from %d", id)
	dc.IncreaseDeauthorizeCount([]string{fmt.Sprintf("%d", id)})

	err := dc.GetUserRepo().DeleteUser(ctx, id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("deauthorize error")
		dc.SendMessage(errors.ErrDeleting.Error(), id, true, true)
		return
	}
	dc.SendMessage("deleted user", id, true, true)
}
