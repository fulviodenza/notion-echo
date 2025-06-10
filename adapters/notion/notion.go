package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"

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
	ID  string `json:"id"`
}

type FileUploadObject struct {
	Object     string `json:"object"`
	ID         string `json:"id"`
	UploadURL  string `json:"upload_url"`
	Status     string `json:"status"`
	ExpiryTime string `json:"expiry_time"`
}

func (ns *Service) UploadFile(ctx context.Context, fileName string, fileData []byte) (*FileUploadResponse, error) {
	fileUpload, err := ns.createFileUpload(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file upload: %w", err)
	}

	err = ns.sendFileContents(ctx, fileUpload.ID, fileName, fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to send file contents: %w", err)
	}

	return &FileUploadResponse{
		ID:  fileUpload.ID,
		URL: fileUpload.UploadURL,
	}, nil
}

func (ns *Service) createFileUpload(ctx context.Context) (*FileUploadObject, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/file_uploads", bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ns.Client.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create file upload: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fileUpload FileUploadObject
	err = json.Unmarshal(body, &fileUpload)
	if err != nil {
		return nil, err
	}

	return &fileUpload, nil
}

func (ns *Service) sendFileContents(ctx context.Context, uploadID, fileName string, fileData []byte) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		ext := filepath.Ext(fileName)
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		case ".txt":
			contentType = "text/plain"
		case ".doc":
			contentType = "application/msword"
		case ".docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		default:
			contentType = "application/octet-stream"
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	h.Set("Content-Type", contentType)

	fw, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = fw.Write(fileData)
	if err != nil {
		return err
	}

	w.Close()

	uploadURL := fmt.Sprintf("https://api.notion.com/v1/file_uploads/%s/send", uploadID)
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ns.Client.Token))
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send file contents: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
