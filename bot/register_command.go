package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	if rc == nil || rc.Bot == nil {
		return
	}

	oauthURL := os.Getenv(OAUTH_URL)

	stateToken, err := generateStateToken()
	if err != nil {
		return
	}

	_, err = rc.Bot.GetUserRepo().SaveUser(ctx, update.Message.Chat.Id, stateToken)
	if err != nil {
		return
	}
	rc.Bot.SetNotionUser(stateToken)

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
