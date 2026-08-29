package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

var _ NotionInterface = (*NotionMock)(nil)

type NotionMock struct {
	pages       map[string]*notionapi.Page
	databases   map[string]*notionapi.Database
	createdPage *notionapi.PageCreateRequest
	err         error
}

func (v *NotionMock) ListPages(ctx context.Context) ([]*notionapi.Page, error) {
	return []*notionapi.Page{
		{
			ID:         "",
			Properties: notionapi.Properties{},
		},
	}, nil
}

func NewNotionMock(pages map[string]*notionapi.Page, err error) NotionInterface {
	return &NotionMock{
		pages: pages,
		err:   err,
	}
}

func NewNotionMockWithDatabases(pages map[string]*notionapi.Page, databases map[string]*notionapi.Database, err error) *NotionMock {
	return &NotionMock{
		pages:     pages,
		databases: databases,
		err:       err,
	}
}

func (v *NotionMock) SearchDatabase(ctx context.Context, databaseName string) ([]*notionapi.Database, error) {
	if v.err != nil {
		return nil, v.err
	}

	databases := []*notionapi.Database{}
	for _, d := range v.databases {
		databases = append(databases, d)
	}

	return databases, nil
}

func (v *NotionMock) CreatePage(ctx context.Context, req *notionapi.PageCreateRequest) (*notionapi.Page, error) {
	if v.err != nil {
		return nil, v.err
	}
	v.createdPage = req
	return &notionapi.Page{
		ID:         "mock-created-page",
		Properties: req.Properties,
	}, nil
}

// CreatedPage exposes the last PageCreateRequest passed to CreatePage so tests
// can assert on the properties the command built.
func (v *NotionMock) CreatedPage() *notionapi.PageCreateRequest {
	return v.createdPage
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

func (v *NotionMock) UploadFile(ctx context.Context, fileName string, fileData []byte) (*FileUploadResponse, error) {
	if v.err != nil {
		return nil, v.err
	}
	return &FileUploadResponse{
		ID:  "mock-file-upload-id-" + fileName,
		URL: "https://notion.so/uploaded/" + fileName,
	}, nil
}
