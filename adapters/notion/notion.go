package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type NotionInterface interface {
	SearchPage(ctx context.Context, pageName string) (*notionapi.Page, error)
	Block() notionapi.BlockService
}

var _ NotionInterface = (*Service)(nil)

type Service struct {
	*notionapi.Client
}

func NewNotionService(client *notionapi.Client) NotionInterface {
	return &Service{
		Client: client,
	}
}

func (ns *Service) SearchPage(ctx context.Context, pageName string) (*notionapi.Page, error) {
	res, err := ns.Search.Do(ctx, &notionapi.SearchRequest{
		Query: pageName,
		Filter: notionapi.SearchFilter{
			Value:    "page",
			Property: "object",
		},
	})
	if err != nil {
		return nil, err
	}

	page := &notionapi.Page{}
	for _, obj := range res.Results {
		page = obj.(*notionapi.Page)
	}

	return page, nil
}

func (ns *Service) Block() notionapi.BlockService {
	return ns.Client.Block
}
