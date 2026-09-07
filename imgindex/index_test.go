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
