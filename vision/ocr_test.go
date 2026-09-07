package vision

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
)

func writeTestPNG(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/test.png"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOCRFile(t *testing.T) {
	path := writeTestPNG(t)
	client := &fakeCompleter{}

	text, err := OCRFile(context.Background(), client, "llava", path)
	if err != nil {
		t.Fatalf("OCRFile() error = %v", err)
	}
	if text != "고양이입니다." {
		t.Errorf("unexpected OCR result: %q", text)
	}
	if len(client.request.Messages) != 1 || len(client.request.Messages[0].ContentParts) != 2 {
		t.Fatalf("unexpected request: %+v", client.request)
	}
	if !strings.Contains(client.request.Messages[0].ContentParts[0].Text, "OCR") {
		t.Errorf("OCR instruction missing: %q", client.request.Messages[0].ContentParts[0].Text)
	}
}
