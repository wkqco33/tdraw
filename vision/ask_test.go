package vision

import (
	"context"
	"encoding/base64"
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
	// http.DetectContentType는 PNM을 감지하지 못하므로 자체 검사로 통과해야 한다.
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
	wantData := base64.StdEncoding.EncodeToString(data)
	if !strings.HasPrefix(parts[1].ImageURL.URL, "data:image/x-portable-graymap;base64,") ||
		!strings.HasSuffix(parts[1].ImageURL.URL, wantData) {
		t.Errorf("unexpected image data URL: %+v", parts[1])
	}
}
