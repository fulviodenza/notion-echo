package bot

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/SakoDroid/telego/v2/objects"
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

func (rc *RegisterCommand) Execute(ctx context.Context, update *objects.Update) {
	id := update.Message.Chat.Id
	rc.Logger().Infof("[RegisterCommand] got registration request from %d", id)

	metrics.RegisterCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	stateToken, err := rc.generateStateToken()
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrStateToken.Error(), id, false, true)
		return
	}
	_, err = rc.GetUserRepo().SaveUser(ctx, update.Message.Chat.Id, stateToken)
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrRegistering.Error(), id, false, true)
		return
	}
	oauthURL := fmt.Sprintf("%s&state=%s", os.Getenv("OAUTH_URL"), url.QueryEscape(stateToken))
	rc.SendMessage("click on the following URL, and authorize the page you want this bot to have access to", id, false, true)
	rc.SendMessage(oauthURL, id, false, false)
	rc.SendMessage("when you have done with registration, select a default page using command /defaultpage with the name of the page you have authorized before", id, true, true)
	rc.Logger().Infof("[RegisterCommand] registration request from %d ended successfully", id)
}
