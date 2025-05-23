package bot

import (
	"context"
	"fmt"
	"net/url"
	"os"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*RegisterCommand)(nil)

type RegisterCommand struct {
	types.IBot
	generateStateToken func() (string, error)
}

func NewRegisterCommand(bot types.IBot, generateStateToken func() (string, error)) types.Command {
	hc := RegisterCommand{
		IBot:               bot,
		generateStateToken: generateStateToken,
	}
	return hc.Execute
}

func (rc *RegisterCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	id := int(update.Message.Chat.ID)
	rc.Logger().Infof("[RegisterCommand] got registration request from %d", id)

	metrics.RegisterCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	stateToken, err := rc.generateStateToken()
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrStateToken.Error(), id, false, true)
		return
	}
	_, err = rc.GetUserRepo().SaveUser(ctx, id, stateToken)
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrRegistering.Error(), id, false, true)
		return
	}
	oauthURL := fmt.Sprintf("%s&state=%s", os.Getenv("OAUTH_URL"), url.QueryEscape(stateToken))
	rc.SendButtonWithURL(update.Message.Chat.ID, "Authorize", oauthURL, "click on the following button, and authorize the page you want this bot to have access to")
	rc.SendMessage("when you have done with registration, select a default page using command /defaultpage with the name of the page you have authorized before", id, true, true)
	rc.Logger().Infof("[RegisterCommand] registration request from %d ended successfully", id)
}
