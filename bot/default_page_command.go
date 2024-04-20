package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	"github.com/notion-echo/errors"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*DefaultPageCommand)(nil)

type DefaultPageCommand struct {
	types.IBot
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error)
}

func NewDefaultPageCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, encKey []byte) (notion.NotionInterface, error)) types.Command {
	hc := DefaultPageCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
	}
	return hc.Execute
}

func (dc *DefaultPageCommand) Execute(ctx context.Context, update *objects.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := update.Message.Chat.Id

	encKey, err := dc.GetVaultClient().GetKey(os.Getenv("VAULT_PATH"))
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		dc.SendMessage(errors.ErrNotRegistered.Error(), id, false, true)
		return
	}
	notionClient, err := dc.buildNotionClient(ctx, dc.GetUserRepo(), update.Message.Chat.Id, encKey)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		dc.SendMessage(errors.ErrSetDefaultPage.Error(), id, false, true)
		return
	}

	selectedPages := strings.Split(update.Message.Text, " ")
	selectedPage := ""
	if len(selectedPages) == 1 {
		dc.SendMessage("please, select a page", id, false, true)
		return
	}
	if len(selectedPages) > 2 {
		dc.SendMessage(fmt.Sprintf("ignoring %v", selectedPages[2:]), id, false, true)
	}
	selectedPage = selectedPages[1]

	p, err := notionClient.SearchPage(ctx, selectedPage)
	if err != nil || p.ID == "" || p.Object == "" {
		if err != nil {
			dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		}
		dc.SendMessage(errors.ErrPageNotFound.Error(), id, false, true)
		return
	}
	err = dc.GetUserRepo().SetDefaultPage(ctx, update.Message.Chat.Id, selectedPage)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		dc.SendMessage(errors.ErrSetDefaultPage.Error(), id, false, true)
		return
	}
	dc.SendMessage(fmt.Sprintf("page %s set as default!", selectedPage), id, false, true)
}
