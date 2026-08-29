package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/notion-echo/adapters/db"
	"github.com/notion-echo/adapters/notion"
	"github.com/notion-echo/bot/types"
	notionerrors "github.com/notion-echo/errors"
	"github.com/notion-echo/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var _ types.ICommand = (*ListDatabasesCommand)(nil)

type ListDatabasesCommand struct {
	types.IBot
	buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)
}

func NewListDatabasesCommand(bot types.IBot, buildNotionClient func(ctx context.Context, userRepo db.UserRepoInterface, id int, token string) (notion.NotionInterface, error)) types.Command {
	hc := ListDatabasesCommand{
		IBot:              bot,
		buildNotionClient: buildNotionClient,
	}
	return hc.Execute
}

func (lc *ListDatabasesCommand) Execute(ctx context.Context, update *tgbotapi.Update) {
	if lc == nil || lc.IBot == nil {
		return
	}

	id := int(update.Message.Chat.ID)
	lc.Logger().Infof("[ListDatabasesCommand] got listdb request from %d", id)

	metrics.ListDbCount.With(prometheus.Labels{"id": fmt.Sprint(id)}).Inc()

	notionToken, err := lc.GetUserRepo().GetNotionTokenByID(ctx, id)
	if err != nil {
		lc.Logger().WithFields(logrus.Fields{"error": err}).Error("listdb error")
		lc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}
	notionClient, err := lc.buildNotionClient(ctx, lc.GetUserRepo(), id, notionToken)
	if err != nil {
		lc.Logger().WithFields(logrus.Fields{"error": err}).Error("listdb error")
		lc.SendMessage(notionerrors.ErrNotRegistered.Error(), id, false, true)
		return
	}

	databases, err := notionClient.SearchDatabase(ctx, "")
	if err != nil {
		lc.Logger().WithFields(logrus.Fields{"error": err}).Error("listdb error")
		lc.SendMessage(notionerrors.ErrDatabaseNotFound.Error(), id, false, true)
		return
	}
	if len(databases) == 0 {
		lc.SendMessage(notionerrors.ErrDatabaseNotFound.Error(), id, false, true)
		return
	}

	names := make([]string, 0, len(databases))
	for _, d := range databases {
		name := notion.ExtractRichText(d.Title)
		if name == "" {
			name = "(untitled)"
		}
		names = append(names, "• "+name)
	}
	sort.Strings(names)

	lc.SendMessage("Databases I can access:\n"+strings.Join(names, "\n"), id, false, false)
}
