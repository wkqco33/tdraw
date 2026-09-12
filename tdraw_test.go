package tdraw

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"testing"
	"time"
)

func writeTempPNG(t *testing.T, w, h int) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "tdraw-test-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestDraw(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.White)
		}
	}

	var buf bytes.Buffer
	Draw(&buf, img, Options{
		Width:     4,
		Height:    4,
		ColorMode: TrueColor,
	})

	out := buf.String()
	if len(out) == 0 {
		t.Fatal("expected non-empty output from Draw")
	}
	if !bytes.Contains(buf.Bytes(), []byte("▀")) {
		t.Error("expected output to contain half-block character ▀")
	}
}

func TestDraw_DefaultOptions(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	// Width 0, Height 0 should auto-detect
	Draw(&buf, img, Options{})
	if buf.Len() == 0 {
		t.Fatal("expected output when default options are passed")
	}
}

func TestDrawFile(t *testing.T) {
	path := writeTempPNG(t, 8, 8)
	var buf bytes.Buffer
	err := DrawFile(&buf, path, Options{
		Width:     8,
		ColorMode: Color256,
	})
	if err != nil {
		t.Fatalf("DrawFile failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output from DrawFile")
	}
}

func TestDrawFile_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := DrawFile(&buf, "/nonexistent/path/file.png", Options{})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestPlayGIFFile_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := PlayGIFFile(context.Background(), &buf, "/nonexistent/anim.gif", Options{})
	if err == nil {
		t.Fatal("expected error for nonexistent GIF")
	}
}

func TestPlayGIFFile_Cancelled(t *testing.T) {
	// Write dummy PNG as fake GIF to test decode error
	path := writeTempPNG(t, 4, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	err := PlayGIFFile(ctx, &buf, path, Options{})
	if err == nil {
		t.Fatal("expected error when decoding non-GIF as GIF")
	}
}

func writeTempGIF(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "tdraw-test-*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{
		Image: []*image.Paletted{
			image.NewPaletted(image.Rect(0, 0, 4, 4), pal),
			image.NewPaletted(image.Rect(0, 0, 4, 4), pal),
		},
		Delay: []int{1, 1},
	}
	if err := gif.EncodeAll(f, g); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestPlayGIFFile_Valid(t *testing.T) {
	path := writeTempGIF(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	err := PlayGIFFile(ctx, &buf, path, Options{
		Width: 4,
	})
	if err != nil {
		t.Fatalf("PlayGIFFile failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output from PlayGIFFile")
	}
}

func TestTermSize(t *testing.T) {
	cols, rows := TermSize()
	if cols <= 0 || rows <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", cols, rows)
	}
}
