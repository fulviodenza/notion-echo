package bot

import (
	"context"
	"fmt"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
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

func (dc *GetDefaultPageCommand) Execute(ctx context.Context, update *objects.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := update.Message.Chat.Id
	dc.Logger().Infof("[GetDefaultPageCommand] got getdefaultpage request from %d", id)
	dc.IncreaseGetDefaultPageCount([]string{fmt.Sprintf("%d", id)})

	defaultPage, err := dc.GetUserRepo().GetDefaultPage(ctx, update.Message.Chat.Id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
	}
	if err != nil || defaultPage == "" {
		dc.SendMessage(errors.ErrPageNotFound.Error(), id, false, true)
		return
	}
	dc.SendMessage(fmt.Sprintf("your default page is **%s**", defaultPage), id, true, true)
}
