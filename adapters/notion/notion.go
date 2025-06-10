package notion

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/jomei/notionapi"
)

type NotionInterface interface {
	SearchPage(ctx context.Context, pageName string) ([]*notionapi.Page, error)
	Block() notionapi.BlockService
	ListPages(ctx context.Context) ([]*notionapi.Page, error)
	UploadFile(ctx context.Context, fileName string, fileData []byte) (*FileUploadResponse, error)
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

func (ns *Service) ListPages(ctx context.Context) ([]*notionapi.Page, error) {
	res, err := ns.Search.Do(ctx, &notionapi.SearchRequest{
		Filter: notionapi.SearchFilter{
			Value:    "page",
			Property: "object",
		},
		Sort: &notionapi.SortObject{
			Direction: "descending",
			Timestamp: "last_edited_time",
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
	if titleProperty, ok := props["title"].(*notionapi.TitleProperty); ok {
		if len(titleProperty.Title) > 0 {
			if titleProperty.Title[0].Text.Content != "" {
				return titleProperty.Title[0].Text.Content
			}
			if titleProperty.Title[0].PlainText != "" {
				return titleProperty.Title[0].PlainText
			}
		}
	}
	return ""
}

type FileUploadResponse struct {
	URL string `json:"url"`
}

func (ns *Service) UploadFile(ctx context.Context, fileName string, fileData []byte) (*FileUploadResponse, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}

	_, err = fw.Write(fileData)
	if err != nil {
		return nil, err
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/files", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ns.Client.Token))
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to upload file: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &FileUploadResponse{
		URL: string(body),
	}, nil
}
