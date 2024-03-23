package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"os"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
)

var _ types.ICommand = (*RegisterCommand)(nil)

type RegisterCommand struct {
	Bot types.IBot
}

func NewRegisterCommand(bot *Bot) types.Command {
	hc := RegisterCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (rc *RegisterCommand) Execute(ctx context.Context, update *objects.Update) {
	redirectURL := os.Getenv(REDIRECT_URL)
	oauthClientID := os.Getenv(OAUTH_CLIENT_ID)

	stateToken, err := generateStateToken()
	if err != nil {
		return
	}

	_, err = rc.Bot.GetUserRepo().SaveUser(ctx, update.Message.Chat.Id, stateToken)
	if err != nil {
		return
	}
	rc.Bot.SetNotionUser(stateToken)

	var oauthURL = "https://api.notion.com/v1/oauth/authorize?" +
		"client_id=" + url.QueryEscape(oauthClientID) +
		"&redirect_uri=" + redirectURL +
		"&state=" + url.QueryEscape(stateToken) +
		"&response_type=code"

	rc.Bot.SendMessage(oauthURL, update, false)
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
