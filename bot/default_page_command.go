package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	notionerrors "github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/notion-echo/utils"
	"github.com/prometheus/client_golang/prometheus"
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
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		dc.SendMessage(notionerrors.ErrTokenNotFound.Error(), id, false, true)
		return
	}
	notionClient, err := dc.buildNotionClient(ctx, dc.GetUserRepo(), id, notionToken)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		dc.SendMessage(notionerrors.ErrSetDefaultPage.Error(), id, false, true)
		return
	}

	selectedPage := ""
	if dc.GetUserState(id) == "" {
		_, selectedPage = utils.SplitFirstOccurrence(update.Message.Text, " ")
		selectedPage = strings.TrimSpace(selectedPage)
	}
	if selectedPage == "" && dc.GetUserState(id) == "" {
		dc.SetUserState(id, "/defaultpage")
		dc.SendMessage("write the page name in the next message", id, false, true)
		return
	}

	if dc.GetUserState(id) != "" {
		selectedPage = strings.TrimSpace(update.Message.Text)
		dc.DeleteUserState(id)
	}

	pages, err := notionClient.SearchPage(ctx, selectedPage)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		return
	}

	if len(pages) == 0 {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("note error")
		dc.SendMessage(notionerrors.ErrBotNotAuthorized.Error(), id, false, true)
		return
	}
	p := pages[0]
	if p.ID == "" || p.Object == "" {
		dc.SendMessage(notionerrors.ErrPageNotFound.Error(), id, false, true)
		return
	}
	err = dc.GetUserRepo().SetDefaultPage(ctx, id, selectedPage)
	if err != nil {
		dc.Logger().WithFields(logrus.Fields{"error": err}).Error("default page error")
		dc.SendMessage(notionerrors.ErrSetDefaultPage.Error(), id, false, true)
		return
	}
	metrics.DefaultPageCount.With(prometheus.Labels{"id": fmt.Sprint(id), "page": string(p.ID)}).Inc()
	dc.SendMessage(fmt.Sprintf("page %s set as default", selectedPage), id, false, true)
}
