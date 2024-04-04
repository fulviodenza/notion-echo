package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/notion-echo/utils"
)

var _ types.ICommand = (*DefaultPageCommand)(nil)

type DefaultPageCommand struct {
	types.IBot
}

func NewDefaultPageCommand(bot *Bot) types.Command {
	hc := DefaultPageCommand{
		IBot: bot,
	}
	return hc.Execute
}

func (dc *DefaultPageCommand) Execute(ctx context.Context, update *objects.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	encKey, err := dc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		dc.SendMessage(errors.ErrNotRegistered.Error(), update, false)
		return
	}
	notionClient, err := buildNotionClient(ctx, dc.GetUserRepo(), update.Message.Chat.Id, encKey)
	if err != nil {
		dc.SendMessage(errors.ErrSetDefaultPage.Error(), update, false)
		return
	}

	selectedPage := strings.Replace(update.Message.Text, utils.COMMAND_DEFAULT_PAGE+" ", "", 1)
	p, err := notionClient.SearchPage(ctx, selectedPage)
	if err != nil || p.ID == "" || p.Object == "" {
		dc.SendMessage(errors.ErrPageNotFound.Error(), update, false)
		return
	}
	err = dc.GetUserRepo().SetDefaultPage(ctx, update.Message.Chat.Id, selectedPage)
	if err != nil {
		log.Println(err)
		dc.SendMessage(errors.ErrSetDefaultPage.Error(), update, false)
		return
	}
	dc.SendMessage(fmt.Sprintf("page %s set as default!", selectedPage), update, false)
}
