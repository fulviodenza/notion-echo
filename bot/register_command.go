package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
)

var _ types.ICommand = (*RegisterCommand)(nil)

type RegisterCommand struct {
	types.IBot
}

func NewRegisterCommand(bot types.IBot) types.Command {
	hc := RegisterCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (rc *RegisterCommand) Execute(ctx context.Context, update *objects.Update) {
	stateToken, err := generateStateToken()
	if err != nil {
		return
	}
	_, err = rc.GetUserRepo().SaveUser(ctx, update.Message.Chat.Id, stateToken)
	if err != nil {
		rc.SendMessage(errors.ErrRegistering.Error(), update, false)
		return
	}
	oauthURL := fmt.Sprintf("%s&state=%s", os.Getenv("OAUTH_URL"), url.QueryEscape(stateToken))
	rc.SendMessage(oauthURL, update, false)
}

func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	stateToken := base64.URLEncoding.EncodeToString(b)
	return stateToken, nil
}
