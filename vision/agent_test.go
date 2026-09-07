package vision

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestImageMetadataTool(t *testing.T) {
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

	result, err := (&imageMetadataTool{path: path}).Execute(context.Background(), "{}")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["format"] != "png" || got["width"] != float64(1) || got["height"] != float64(1) {
		t.Errorf("unexpected metadata: %s", result)
	}
}
