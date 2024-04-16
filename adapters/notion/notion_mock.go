package notion

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/notion-echo/errors"
)

var _ NotionInterface = (*NotionMock)(nil)

type NotionMock struct {
	pages map[string]*notionapi.Page
	err   error
}

func NewNotionMock(pages map[string]*notionapi.Page, err error) NotionInterface {
	return &NotionMock{
		pages: pages,
		err:   err,
	}
}

func (v *NotionMock) SearchPage(ctx context.Context, pageName string) (*notionapi.Page, error) {
	if v.err != nil {
		return nil, v.err
	}
	p, ok := v.pages[pageName]
	if !ok {
		return nil, errors.ErrPageNotFound
	}
	return p, nil
}

func (v *NotionMock) Block() notionapi.BlockService {
	return nil
}
