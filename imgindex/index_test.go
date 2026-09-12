package imgindex

import (
	"context"
	"os"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
)

type fakeCompleter struct{}

func (fakeCompleter) Complete(_ context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{Choices: []llm.Choice{{Message: llm.Message{Content: "터미널 오류 화면"}}}}, nil
}

func TestIndexSearchAndRoundTrip(t *testing.T) {
	idx := &Index{
		Version: 1,
		Root:    "photos",
		Entries: []Entry{
			{Path: "photos/error.png", Description: "터미널 오류 화면"},
			{Path: "photos/cat.jpg", Description: "소파 위의 고양이"},
		},
	}
	path := t.TempDir() + "/nested/index.json"
	if err := idx.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	results := loaded.Search("고양이", 10)
	if len(results) != 1 || results[0].Path != "photos/cat.jpg" {
		t.Fatalf("unexpected search results: %+v", results)
	}
}

func TestBuildSkipsIndexDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root+"/.tdraw", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root+"/.tdraw/ignored.png", []byte("not used"), 0600); err != nil {
		t.Fatal(err)
	}
	idx, err := Build(context.Background(), root, fakeCompleter{}, "llava", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Entries) != 0 {
		t.Fatalf("expected empty index, got %+v", idx.Entries)
	}
}

func TestBuildWithRealImages(t *testing.T) {
	root := t.TempDir()
	pngPath := root + "/test.png"
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Write simple 1x1 PNG
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG header
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, // IDAT
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, // IEND
		0x42, 0x60, 0x82,
	}
	if _, err := f.Write(pngData); err != nil {
		t.Fatal(err)
	}
	f.Close()

	idx, err := Build(context.Background(), root, fakeCompleter{}, "llava", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(idx.Entries))
	}
	if idx.Entries[0].Width != 1 || idx.Entries[0].Height != 1 || idx.Entries[0].Format != "png" {
		t.Errorf("unexpected entry: %+v", idx.Entries[0])
	}
}
