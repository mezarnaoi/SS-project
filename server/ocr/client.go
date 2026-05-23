package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

const defaultOCRTimeout = 35 * time.Second

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Result struct {
	Text                string  `json:"text"`
	Confidence          float64 `json:"confidence"`
	RecognizedWordCount int     `json:"recognized_word_count"`
	Engine              string  `json:"engine"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultOCRTimeout,
		},
	}
}

func (c *Client) ExtractText(ctx context.Context, imageData []byte, imageType string) (string, float64, error) {
	if len(imageData) == 0 {
		return "", 0, fmt.Errorf("empty image payload")
	}

	contentType, filename := contentTypeForImage(imageType)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	partHeader.Set("Content-Type", contentType)

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create OCR multipart file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return "", 0, fmt.Errorf("failed to write OCR multipart file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to close OCR multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/ocr", &requestBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to build OCR request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("OCR sandbox request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", 0, fmt.Errorf("failed to read OCR response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("OCR sandbox returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	var result Result
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", 0, fmt.Errorf("failed to decode OCR response: %w", err)
	}

	return result.Text, result.Confidence, nil
}

func contentTypeForImage(imageType string) (string, string) {
	switch imageType {
	case "png":
		return "image/png", "image.png"
	case "jpeg", "jpg":
		return "image/jpeg", "image.jpg"
	default:
		return "application/octet-stream", "image.bin"
	}
}
