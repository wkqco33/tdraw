// Package imgindex builds and searches a local image analysis index.
package imgindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wkqco33/tdraw/vision"
)

// Entry is one indexed image.
type Entry struct {
	Path        string `json:"path"`
	Format      string `json:"format"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Description string `json:"description"`
}

// Index is the persisted image index.
type Index struct {
	Version int     `json:"version"`
	Root    string  `json:"root"`
	Entries []Entry `json:"entries"`
}

// Build scans root and asks the Vision model for one concise description per image.
func Build(ctx context.Context, root string, client vision.Completer, model string, progress func(string)) (*Index, error) {
	idx := &Index{Version: 1, Root: root}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".tdraw" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isImagePath(path) {
			return nil
		}

		info, err := loadImage(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if progress != nil {
			progress(path)
		}
		description, err := vision.AskFile(ctx, client, model, path,
			"이 이미지를 한 문장으로 설명해줘. 보이는 주요 객체, 장면, 문서 종류를 포함하고 설명만 출력해줘.")
		if err != nil {
			return err
		}
		idx.Entries = append(idx.Entries, Entry{
			Path:        path,
			Format:      info.Format,
			Width:       info.Width,
			Height:      info.Height,
			Description: strings.TrimSpace(description),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(idx.Entries, func(i, j int) bool { return idx.Entries[i].Path < idx.Entries[j].Path })
	return idx, nil
}

// Save writes an index as indented JSON and creates its parent directory.
func (idx *Index) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

// Load reads an index from disk.
func Load(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("인덱스 JSON 해석 실패: %w", err)
	}
	return &idx, nil
}

// Search returns entries ranked by simple local token matching.
func (idx *Index) Search(query string, limit int) []Entry {
	terms := strings.Fields(strings.ToLower(query))
	type scored struct {
		entry Entry
		score int
	}
	results := make([]scored, 0, len(idx.Entries))
	for _, entry := range idx.Entries {
		text := strings.ToLower(entry.Path + " " + entry.Description)
		score := 0
		for _, term := range terms {
			if strings.Contains(text, term) {
				score++
			}
		}
		if score > 0 {
			results = append(results, scored{entry: entry, score: score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].entry.Path < results[j].entry.Path
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	out := make([]Entry, len(results))
	for i, result := range results {
		out[i] = result.entry
	}
	return out
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".pgm", ".ppm", ".pbm":
		return true
	default:
		return false
	}
}

func loadImage(path string) (*imageInfo, error) {
	// Keep the index package independent from image.Image storage.
	info, err := load(path)
	if err != nil {
		return nil, err
	}
	return &imageInfo{Format: info.format, Width: info.width, Height: info.height}, nil
}

type imageInfo struct {
	Format        string
	Width, Height int
}
