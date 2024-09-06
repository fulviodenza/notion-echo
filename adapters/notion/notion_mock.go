package notion

import (
	"context"

	"github.com/jomei/notionapi"
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

func (v *NotionMock) SearchPage(ctx context.Context, pageName string) ([]*notionapi.Page, error) {
	if v.err != nil {
		return nil, v.err
	}

	pages := []*notionapi.Page{}
	for _, v := range v.pages {
		pages = append(pages, v)
	}

	return pages, nil
}

func (v *NotionMock) Block() notionapi.BlockService {
	return &BlockServiceMock{}
}

type BlockService interface {
	notionapi.BlockService
}

type BlockServiceMock struct{}

func (bsm BlockServiceMock) AppendChildren(context.Context, notionapi.BlockID, *notionapi.AppendBlockChildrenRequest) (*notionapi.AppendBlockChildrenResponse, error) {
	return nil, nil
}

func (bsm BlockServiceMock) Get(context.Context, notionapi.BlockID) (notionapi.Block, error) {
	return &notionapi.CalloutBlock{}, nil
}
func (bsm BlockServiceMock) GetChildren(context.Context, notionapi.BlockID, *notionapi.Pagination) (*notionapi.GetChildrenResponse, error) {
	return nil, nil
}
func (bsm BlockServiceMock) Update(ctx context.Context, id notionapi.BlockID, request *notionapi.BlockUpdateRequest) (notionapi.Block, error) {
	return &notionapi.CalloutBlock{}, nil
}
func (bsm BlockServiceMock) Delete(context.Context, notionapi.BlockID) (notionapi.Block, error) {
	return &notionapi.CalloutBlock{}, nil
}
