package broker_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"mqtt-streaming-server/ocr"
)

// FuzzImageDecode verifies image decoding never panics with arbitrary input.
// Relevant for security: malicious images sent via MQTT could crash the server.
func FuzzImageDecode(f *testing.F) {
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10})             // JPEG header
	f.Add([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) // PNG header
	f.Add([]byte("not an image"))
	f.Add([]byte{0x00, 0x00, 0x00, 0x00})
	f.Add([]byte{})
	f.Add([]byte("<script>alert(1)</script>"))
	f.Add(bytes.Repeat([]byte{0xFF}, 1024))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = image.DecodeConfig(bytes.NewReader(data))
	})
}

// FuzzOCRClient verifies that the Go OCR sandbox client handles arbitrary image
// bytes without panicking. The real OCR engine runs in the isolated ocr-sandbox
// container, so this fuzz test uses a local HTTP test server instead of calling
// Tesseract directly.
func FuzzOCRClient(f *testing.F) {
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE0}, "jpeg")
	f.Add([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "png")
	f.Add([]byte("not an image"), "jpeg")
	f.Add([]byte{0x00}, "png")
	f.Add([]byte{}, "jpeg")
	f.Add(bytes.Repeat([]byte{0xAB}, 512), "jpeg")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ocr" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(8 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		var received bytes.Buffer
		_, _ = received.ReadFrom(file)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text":                  "mock OCR text",
			"confidence":            87.5,
			"recognized_word_count": 3,
			"engine":                "mock-ocr-sandbox",
		})
	}))
	defer server.Close()

	ocrClient := ocr.NewClient(server.URL)

	f.Fuzz(func(t *testing.T, data []byte, imageType string) {
		_, _, _ = ocrClient.ExtractText(context.Background(), data, imageType)
	})
}
