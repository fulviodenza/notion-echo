package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/prometheus/client_golang/prometheus"
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

func (dc *DeauthorizeCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
	dc.Logger().Infof("[DeauthorizeCommand] got deauthorize request from %d", id)

	metrics.DeauthorizeCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	err := dc.GetUserRepo().DeleteUser(ctx, id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("deauthorize error")
		dc.SendMessage(errors.ErrDeleting.Error(), id, true, true)
		return
	}
	dc.SendMessage("deleted user", id, true, true)
}
