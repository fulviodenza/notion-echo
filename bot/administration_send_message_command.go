package bot

import (
	"context"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*SendAllCommand)(nil)

type SendAllCommand struct {
	types.IBot
}

func NewSendAllCommand(bot types.IBot) types.Command {
	hc := SendAllCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (sa *SendAllCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if sa == nil || sa.IBot == nil {
		return
	}
	id := int(update.Message.Chat.ID)

	tg_id, err := strconv.Atoi(os.Getenv("TG_ID"))
	if err != nil {
		sa.Logger().WithFields(logrus.Fields{"error": err}).Error("send all")
		sa.SendMessage(errors.ErrDeleting.Error(), id, true, true)
		return
	}
	if id != tg_id {
		return
	}
	users, err := sa.GetUserRepo().GetAllUsers(ctx)
	if err != nil {
		sa.Logger().WithFields(logrus.Fields{"error": err}).Error("send all")
		sa.SendMessage(errors.ErrDeleting.Error(), id, true, true)
		return
	}

	sendText := strings.Replace(update.Message.Text, "/send_all", "", 1)
	if sendText == "" && update.Message.Text != "" {
		sa.SendMessage("write something in your send_all message", id, false, true)
		return
	}
	for _, u := range users {
		sa.SendMessage(sendText, u.ID, true, true)
	}
	sa.SendMessage("message sent to all", id, false, true)
}
