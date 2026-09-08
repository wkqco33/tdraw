package imgutil

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// --- LoadGIF 테스트 ---

// writeTempGIF는 nFrames 프레임짜리 임시 애니메이션 GIF를 만든다.
// i번째 프레임은 팔레트 인덱스 (i%3)+1 단색으로 칠해진다.
func writeTempGIF(t *testing.T, w, h, nFrames int, disposal byte) string {
	t.Helper()
	pal := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}
	g := &gif.GIF{LoopCount: 0}
	for i := 0; i < nFrames; i++ {
		img := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		ci := uint8(i%3 + 1)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.SetColorIndex(x, y, ci)
			}
		}
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 10) // 0.1s
		g.Disposal = append(g.Disposal, disposal)
	}
	f, err := os.CreateTemp(t.TempDir(), "test-*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, g); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestLoadGIF_Frames(t *testing.T) {
	path := writeTempGIF(t, 32, 24, 4, gif.DisposalNone)
	anim, err := LoadGIF(path)
	if err != nil {
		t.Fatalf("GIF 로드 실패: %v", err)
	}
	if len(anim.Frames) != 4 {
		t.Errorf("프레임 수: got %d, want 4", len(anim.Frames))
	}
	if len(anim.Delays) != 4 {
		t.Errorf("delay 수: got %d, want 4", len(anim.Delays))
	}
	if anim.Width != 32 || anim.Height != 24 {
		t.Errorf("캔버스 크기: got %dx%d, want 32x24", anim.Width, anim.Height)
	}
	// 각 프레임은 전체 캔버스 크기로 합성되어야 한다
	for i, f := range anim.Frames {
		b := f.Bounds()
		if b.Dx() != 32 || b.Dy() != 24 {
			t.Errorf("프레임 %d 크기: got %dx%d, want 32x24", i, b.Dx(), b.Dy())
		}
	}
}

func TestLoadGIF_DelayConversion(t *testing.T) {
	// delay 10(1/100초) → 100ms 로 변환되어야 한다
	path := writeTempGIF(t, 16, 16, 2, gif.DisposalNone)
	anim, err := LoadGIF(path)
	if err != nil {
		t.Fatalf("GIF 로드 실패: %v", err)
	}
	for i, d := range anim.Delays {
		if d != 100*time.Millisecond {
			t.Errorf("프레임 %d delay: got %v, want 100ms", i, d)
		}
	}
}

func TestLoadGIF_Composited(t *testing.T) {
	// 단색 프레임이므로 합성 결과 (0,0) 픽셀이 팔레트 색과 일치해야 한다
	path := writeTempGIF(t, 8, 8, 3, gif.DisposalNone)
	anim, err := LoadGIF(path)
	if err != nil {
		t.Fatalf("GIF 로드 실패: %v", err)
	}
	want := []color.RGBA{
		{255, 0, 0, 255}, // ci=1
		{0, 255, 0, 255}, // ci=2
		{0, 0, 255, 255}, // ci=3
	}
	for i, f := range anim.Frames {
		r, g, b, a := f.At(0, 0).RGBA()
		got := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		if got != want[i] {
			t.Errorf("프레임 %d 색: got %v, want %v", i, got, want[i])
		}
	}
}

func TestLoadGIF_NotFound(t *testing.T) {
	_, err := LoadGIF("/nonexistent/path/anim.gif")
	if err == nil {
		t.Error("없는 파일에 대해 에러가 발생해야 함")
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

// writeTempPGM은 2x2 회색조 PGM(P5 바이너리) 파일을 생성하고 경로를 반환한다.
func writeTempPGM(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.pgm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("P5\n2 2\n255\n\x00\x55\xaa\xff"); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestLoad_PGM(t *testing.T) {
	path := writeTempPGM(t)
	info, err := Load(path)
	if err != nil {
		t.Fatalf("PGM 로드 실패: %v", err)
	}
	if info.Width != 2 || info.Height != 2 {
		t.Errorf("크기: got %dx%d, want 2x2", info.Width, info.Height)
	}
	if info.Format != "pgm" {
		t.Errorf("Format: got %s, want pgm", info.Format)
	}
	// 픽셀 값 검증 (회색조)
	want := []uint8{0, 85, 170, 255}
	for i, v := range want {
		x, y := i%2, i/2
		if got := info.Image.At(x, y).(color.Gray).Y; got != v {
			t.Errorf("픽셀(%d,%d): got %d, want %d", x, y, got, v)
		}
	}
}
