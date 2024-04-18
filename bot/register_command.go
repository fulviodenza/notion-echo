package bot

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
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
	stateToken, err := rc.generateStateToken()
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrStateToken.Error(), update, false)
		return
	}
	_, err = rc.GetUserRepo().SaveUser(ctx, update.Message.Chat.Id, stateToken)
	if err != nil {
		rc.Logger().WithFields(logrus.Fields{"error": err}).Error("register error")
		rc.SendMessage(errors.ErrRegistering.Error(), update, false)
		return
	}
	oauthURL := fmt.Sprintf("%s&state=%s", os.Getenv("OAUTH_URL"), url.QueryEscape(stateToken))
	rc.SendMessage("click on the following URL, authorize pages", update, false)
	rc.SendMessage(oauthURL, update, false)
	rc.SendMessage("when you have done with registration, select a default page using command `/defaultpage page`", update, true)
}
