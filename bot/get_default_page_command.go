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

var _ types.ICommand = (*GetDefaultPageCommand)(nil)

type GetDefaultPageCommand struct {
	types.IBot
}

func NewGetDefaultPageCommand(bot types.IBot) types.Command {
	hc := GetDefaultPageCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (dc *GetDefaultPageCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
	dc.Logger().Infof("[GetDefaultPageCommand] got getdefaultpage request from %d", id)

	metrics.GetDefaultPageCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	defaultPage, err := dc.GetUserRepo().GetDefaultPage(ctx, id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
	}
	if err != nil || defaultPage == "" {
		dc.SendMessage(errors.ErrPageNotFound.Error(), id, false, true)
		return
	}
	dc.SendMessage(fmt.Sprintf("your default page is *bold *%s*", defaultPage), id, true, true)
}
