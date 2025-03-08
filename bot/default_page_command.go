package bot

import (
	"context"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	notionerrors "github.com/notion-echo/errors"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*DefaultPageCommand)(nil)

type DefaultPageCommand struct {
	types.IBot
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, notionToken string) (notion.NotionInterface, error)
}

func NewDefaultPageCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, notionToken string) (notion.NotionInterface, error)) types.Command {
	hc := DefaultPageCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
	}
	return hc.Execute
}

func (dc *DefaultPageCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if dc == nil || dc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
	dc.Logger().Infof("[DefaultPageCommand] got defaultpage request from %d", id)

	notionToken, err := dc.GetUserRepo().GetNotionTokenByID(ctx, id)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Errorf("default page error: %v", err)
		dc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}
	notionClient, err := dc.buildNotionClient(ctx, dc.GetUserRepo(), id, notionToken)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Errorf("default page error: %v", err)
		dc.SendMessage(notionerrors.ErrSetDefaultPage.Error(), id, false, true)
		return
	}

	pages, err := notionClient.ListPages(ctx)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Errorf("default page error: %v", err)
		return
	}
	if len(pages) == 0 {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Errorf("default page error: %v", err)
		dc.SendMessage(notionerrors.ErrBotNotAuthorized.Error(), id, false, true)
		return
	}

	err = dc.SendButtonWithData(
		int64(id),
		"Select the page you want to set as default",
		pages,
	)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Errorf("default page error: %v", err)
	}
}
