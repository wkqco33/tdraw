package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"os"
	"strings"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
)

type fakeCompleter struct {
	request llm.ChatRequest
}

func (f *fakeCompleter) Complete(_ context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	f.request = req
	return &llm.ChatResponse{Choices: []llm.Choice{
		{Message: llm.Message{Content: "고양이입니다."}},
	}}, nil
}

func TestAskFile(t *testing.T) {
	path := t.TempDir() + "/test.png"
	data := []byte{137, 80, 78, 71, 13, 10, 26, 10, 't', 'e', 's', 't'}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	client := &fakeCompleter{}
	answer, err := AskFile(context.Background(), client, "llava", path, "무엇이 보이나요?")
	if err != nil {
		t.Fatalf("AskFile() error = %v", err)
	}
	if answer != "고양이입니다." {
		t.Errorf("unexpected answer: %q", answer)
	}

	if client.request.Model != "llava" || len(client.request.Messages) != 1 {
		t.Fatalf("unexpected request: %+v", client.request)
	}
	parts := client.request.Messages[0].ContentParts
	if len(parts) != 2 || parts[0].Text != "무엇이 보이나요?" {
		t.Fatalf("unexpected content parts: %+v", parts)
	}
	wantData := base64.StdEncoding.EncodeToString(data)
	if parts[1].ImageURL == nil || !strings.HasPrefix(parts[1].ImageURL.URL, "data:image/") ||
		!strings.HasSuffix(parts[1].ImageURL.URL, wantData) {
		t.Errorf("unexpected image data URL: %+v", parts[1])
	}
}

func TestAskFile_RejectsNonImage(t *testing.T) {
	path := t.TempDir() + "/test.txt"
	if err := os.WriteFile(path, []byte("not an image"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := AskFile(context.Background(), &fakeCompleter{}, "llava", path, "분석해줘")
	if err == nil || !strings.Contains(err.Error(), "이미지 파일이 아닙니다") {
		t.Fatalf("expected image validation error, got %v", err)
	}
}

func TestAskFile_PGM(t *testing.T) {
	path := t.TempDir() + "/test.pgm"
	data := []byte("P5\n2 2\n255\n\x00\x55\xaa\xff")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	client := &fakeCompleter{}
	if _, err := AskFile(context.Background(), client, "llava", path, "분석해줘"); err != nil {
		t.Fatalf("AskFile() error = %v", err)
	}

	parts := client.request.Messages[0].ContentParts
	if len(parts) != 2 || parts[1].ImageURL == nil {
		t.Fatalf("unexpected content parts: %+v", parts)
	}
	if !strings.HasPrefix(parts[1].ImageURL.URL, "data:image/png;base64,") {
		t.Errorf("unexpected image data URL: %+v", parts[1])
	}

	encoded, ok := strings.CutPrefix(parts[1].ImageURL.URL, "data:image/png;base64,")
	if !ok {
		t.Fatalf("missing PNG data URL prefix: %q", parts[1].ImageURL.URL)
	}
	pngData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode PNG data URL: %v", err)
	}
	decoded, format, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("decode converted PNG: %v", err)
	}
	if format != "png" {
		t.Fatalf("converted format = %q, want png", format)
	}
	if got := decoded.Bounds().Size(); got != image.Pt(2, 2) {
		t.Fatalf("converted dimensions = %v, want 2x2", got)
	}
}

func TestFileMessage_AllPNMFormatsBecomePNG(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "P1", data: []byte("P1\n2 1\n0 1\n")},
		{name: "P2", data: []byte("P2\n2 1\n255\n0 255\n")},
		{name: "P3", data: []byte("P3\n2 1\n255\n255 0 0 0 255 0\n")},
		{name: "P4", data: []byte("P4\n2 1\n\x40")},
		{name: "P5", data: []byte("P5\n2 1\n255\n\x00\xff")},
		{name: "P6", data: []byte("P6\n2 1\n255\n\xff\x00\x00\x00\xff\x00")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := t.TempDir() + ".pnm"
			if err := os.WriteFile(path, tt.data, 0600); err != nil {
				t.Fatal(err)
			}

			message, err := FileMessage(path, "분석해줘")
			if err != nil {
				t.Fatalf("FileMessage() error = %v", err)
			}
			if len(message.ContentParts) != 2 || message.ContentParts[1].ImageURL == nil {
				t.Fatalf("unexpected content parts: %+v", message.ContentParts)
			}

			encoded, ok := strings.CutPrefix(message.ContentParts[1].ImageURL.URL, "data:image/png;base64,")
			if !ok {
				t.Fatalf("PNM was not normalized to PNG: %q", message.ContentParts[1].ImageURL.URL[:min(40, len(message.ContentParts[1].ImageURL.URL))])
			}
			pngData, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("decode PNG data URL: %v", err)
			}
			img, format, err := image.Decode(bytes.NewReader(pngData))
			if err != nil || format != "png" || img.Bounds().Dx() != 2 || img.Bounds().Dy() != 1 {
				t.Fatalf("converted image = format %q, bounds %v, error %v", format, img.Bounds(), err)
			}
		})
	}
}

func TestAskFile_InvalidPNM(t *testing.T) {
	path := t.TempDir() + "/invalid.pgm"
	if err := os.WriteFile(path, []byte("P2\n2 2\n255\n0"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := AskFile(context.Background(), &fakeCompleter{}, "llava", path, "분석해줘")
	if err == nil || !strings.Contains(err.Error(), "PNM 이미지 디코딩 실패") {
		t.Fatalf("expected PNM decoding error, got %v", err)
	}
}
