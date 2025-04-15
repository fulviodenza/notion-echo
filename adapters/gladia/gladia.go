package gladia

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	gladiaclient "github.com/fulviodenza/go-gladia-client/pkg/gladia"
)

func HandleTranscribe(ctx context.Context, bot *tgbotapi.BotAPI, voice *tgbotapi.Voice) (string, error) {
	fileConfig := tgbotapi.FileConfig{
		FileID: voice.FileID,
	}

	file, err := bot.GetFile(fileConfig)
	if err != nil {
		log.Printf("Failed to get file: %v", err)
		return "", err
	}

	fileURL := file.Link(bot.Token)
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "voice_"+voice.FileID+".ogg")

	err = downloadFile(fileURL, tempFile)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		return "", err
	}
	defer os.Remove(tempFile)

	client := gladiaclient.NewClient(os.Getenv("GLADIA_API_KEY"))

	result, err := client.UploadFile(ctx, tempFile)
	if err != nil {
		log.Printf("Failed to upload file to Gladia: %v", err)
		return "", err
	}

	transcribeErr := client.Transcribe(ctx, result.AudioURL)
	if transcribeErr != nil {
		log.Printf("Failed to start transcription: %v", transcribeErr)
		return "", transcribeErr
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	resultCh := make(chan string)
	errCh := make(chan error)

	// Start a goroutine to poll for results
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeoutCtx.Done():
				errCh <- fmt.Errorf("timeout reached waiting for transcription")
				return
			case <-ticker.C:
				// Poll for results from Gladia client
				transcriptionResult, pollErr := client.GetTranscriptionResult(ctx, result.AudioMetadata.ID)
				if pollErr != nil {
					log.Printf("Error polling for result: %v", pollErr)
					continue
				}

				if transcriptionResult.Status == "completed" {
					resultCh <- transcriptionResult.Result.Transcription.FullTranscript
					return
				} else if transcriptionResult.Status == "error" {
					errCh <- fmt.Errorf("transcription failed: %d", transcriptionResult.ErrorCode)
					return
				}

				log.Printf("Transcription in progress, status: %s", transcriptionResult.Status)
			}
		}
	}()

	// Wait for either a result or an error
	select {
	case transcript := <-resultCh:
		return transcript, nil
	case err := <-errCh:
		return "", err
	case <-timeoutCtx.Done():
		return "", fmt.Errorf("transcription timed out")
	}
}

// download a file from URL to a local path
func downloadFile(url string, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
