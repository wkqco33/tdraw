package imgutil

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// newTestImage는 테스트용 단색 RGBA 이미지를 반환한다.
func newTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// writeTempPNG는 임시 PNG 파일을 생성하고 경로를 반환한다.
func writeTempPNG(t *testing.T, w, h int) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := newTestImage(w, h, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

// --- Resize 테스트 ---

func TestResize_AspectRatio_WidthLimited(t *testing.T) {
	src := newTestImage(200, 100, color.White)
	// targetW=100, targetH=1000 → 폭이 제약
	out := Resize(src, 100, 1000)
	b := out.Bounds()
	// 스케일 0.5: 결과 100x50
	if b.Dx() != 100 {
		t.Errorf("폭: got %d, want 100", b.Dx())
	}
	if b.Dy() != 50 {
		t.Errorf("높이: got %d, want 50", b.Dy())
	}
}

func TestResize_AspectRatio_HeightLimited(t *testing.T) {
	src := newTestImage(200, 100, color.White)
	// targetW=1000, targetH=40 → 높이 40은 짝수이므로 스케일 0.4: 80x40
	out := Resize(src, 1000, 40)
	b := out.Bounds()
	if b.Dx() != 80 {
		t.Errorf("폭: got %d, want 80", b.Dx())
	}
	if b.Dy() != 40 {
		t.Errorf("높이: got %d, want 40", b.Dy())
	}
}

func TestResize_EvenHeight(t *testing.T) {
	// 100x101 이미지를 리사이즈하면 높이가 짝수여야 한다
	src := newTestImage(100, 101, color.White)
	out := Resize(src, 50, 200)
	h := out.Bounds().Dy()
	if h%2 != 0 {
		t.Errorf("높이가 홀수: %d", h)
	}
}

func TestResize_MinSize(t *testing.T) {
	src := newTestImage(1000, 1000, color.White)
	// 아주 작은 목표 크기
	out := Resize(src, 1, 2)
	b := out.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		t.Errorf("크기가 0 이하: %dx%d", b.Dx(), b.Dy())
	}
}

func TestResize_OutputFitsTarget(t *testing.T) {
	src := newTestImage(320, 240, color.White)
	out := Resize(src, 80, 60)
	b := out.Bounds()
	if b.Dx() > 80 {
		t.Errorf("폭이 targetW 초과: got %d", b.Dx())
	}
	if b.Dy() > 60 {
		t.Errorf("높이가 targetH 초과: got %d", b.Dy())
	}
}

// --- Load 테스트 ---

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/image.png")
	if err == nil {
		t.Error("없는 파일에 대해 에러가 발생해야 함")
	}
}

func TestLoad_PNG(t *testing.T) {
	path := writeTempPNG(t, 64, 48)
	info, err := Load(path)
	if err != nil {
		t.Fatalf("PNG 로드 실패: %v", err)
	}
	if info.Width != 64 {
		t.Errorf("Width: got %d, want 64", info.Width)
	}
	if info.Height != 48 {
		t.Errorf("Height: got %d, want 48", info.Height)
	}
	if info.Format != "png" {
		t.Errorf("Format: got %s, want png", info.Format)
	}
	if filepath.Base(info.Path) != filepath.Base(path) {
		t.Errorf("Path: got %s", info.Path)
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-*.png")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("not an image")
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Error("잘못된 파일에 대해 에러가 발생해야 함")
	}
}
