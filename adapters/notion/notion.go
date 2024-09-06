package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type NotionInterface interface {
	SearchPage(ctx context.Context, pageName string) ([]*notionapi.Page, error)
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

func (ns *Service) SearchPage(ctx context.Context, pageName string) ([]*notionapi.Page, error) {
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

	pages := []*notionapi.Page{}
	for _, obj := range res.Results {
		pages = append(pages, obj.(*notionapi.Page))
	}

	return pages, nil
}

func (ns *Service) Block() notionapi.BlockService {
	return ns.Client.Block
}

type NotionPageName struct {
	Title       string   `json:"title,omitempty"`
	Select      string   `json:"select,omitempty"`
	MultiSelect []string `json:"multi_select,omitempty"`
	Status      string   `json:"status,omitempty"`
}

func ExtractName(props notionapi.Properties) string {
	if titleProperty, ok := props["Title"].(notionapi.TitleProperty); ok {
		if len(titleProperty.Title) > 0 {
			return titleProperty.Title[0].Text.Content
		}
	}
	return ""
}
