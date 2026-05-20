package broker_test

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	"github.com/otiai10/gosseract/v2"
)

// FuzzImageDecode verifies image decoding never panics with arbitrary input.
// Relevant for security: malicious images sent via MQTT could crash the server.
func FuzzImageDecode(f *testing.F) {
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}) // JPEG header
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

// FuzzOCRProcessing verifies Tesseract never panics with arbitrary image bytes.
// Relevant for security: prevents RCE via maliciously crafted images (CVE class).
func FuzzOCRProcessing(f *testing.F) {
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	f.Add([]byte("not an image"))
	f.Add([]byte{0x00})
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xAB}, 512))

	f.Fuzz(func(t *testing.T, data []byte) {
		client := gosseract.NewClient()
		defer client.Close()

		_ = client.SetImageFromBytes(data)   // #nosec G104
		_, _ = client.Text()
	})
}
