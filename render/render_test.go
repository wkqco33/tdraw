package render

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// newSolidImage는 w×h 단색 이미지를 반환한다.
func newSolidImage(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// --- Render 출력 구조 테스트 ---

func TestRender_LineCount(t *testing.T) {
	// 높이 20 → 렌더 출력 행 = 20/2 = 10
	img := newSolidImage(10, 20, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	var buf strings.Builder
	Render(&buf, img, ColorTruecolor)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 10 {
		t.Errorf("행 수: got %d, want 10", len(lines))
	}
}

func TestRender_ContainsHalfBlock(t *testing.T) {
	img := newSolidImage(5, 4, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var buf strings.Builder
	Render(&buf, img, ColorTruecolor)
	if !strings.Contains(buf.String(), "▀") {
		t.Error("출력에 half-block 문자(▀)가 없음")
	}
}

func TestRender_ContainsReset(t *testing.T) {
	img := newSolidImage(4, 4, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	var buf strings.Builder
	Render(&buf, img, ColorTruecolor)
	if !strings.Contains(buf.String(), "\033[0m") {
		t.Error("행 끝에 ANSI 리셋 시퀀스가 없음")
	}
}

func TestRender_TruecolorSequence(t *testing.T) {
	img := newSolidImage(2, 2, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	var buf strings.Builder
	Render(&buf, img, ColorTruecolor)
	out := buf.String()
	// truecolor: ESC[38;2;R;G;Bm 형식
	if !strings.Contains(out, "\033[38;2;") {
		t.Error("truecolor fg 시퀀스(ESC[38;2;)가 없음")
	}
	if !strings.Contains(out, "\033[48;2;") {
		t.Error("truecolor bg 시퀀스(ESC[48;2;)가 없음")
	}
}

func TestRender_256Sequence(t *testing.T) {
	img := newSolidImage(2, 2, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	var buf strings.Builder
	Render(&buf, img, Color256)
	out := buf.String()
	if !strings.Contains(out, "\033[38;5;") {
		t.Error("256색 fg 시퀀스(ESC[38;5;)가 없음")
	}
	if !strings.Contains(out, "\033[48;5;") {
		t.Error("256색 bg 시퀀스(ESC[48;5;)가 없음")
	}
	// truecolor 시퀀스가 섞여 있으면 안 됨
	if strings.Contains(out, "\033[38;2;") {
		t.Error("256 모드인데 truecolor 시퀀스가 포함됨")
	}
}

func TestRender_GraySequence(t *testing.T) {
	img := newSolidImage(2, 2, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	var buf strings.Builder
	Render(&buf, img, ColorGray)
	out := buf.String()
	if !strings.Contains(out, "\033[38;5;") {
		t.Error("gray 모드 fg 시퀀스(ESC[38;5;)가 없음")
	}
}

// wrapImage는 *image.RGBA를 감싸 타입 단언을 실패시켜 RenderString의 폴백(At) 경로를
// 강제한다. 빠른 경로(RGBAAt)와 출력이 바이트 단위로 같아야 한다.
type wrapImage struct{ image.Image }

func TestRenderString_FastPathMatchesFallback(t *testing.T) {
	// 픽셀마다 색이 다른 그라데이션(불투명)으로 두 경로를 비교
	img := image.NewRGBA(image.Rect(0, 0, 12, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 12; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 20), G: uint8(y * 30), B: uint8(x ^ y), A: 255})
		}
	}
	for _, mode := range []ColorMode{ColorTruecolor, Color256, ColorGray} {
		fast := RenderString(img, mode)
		slow := RenderString(wrapImage{img}, mode)
		if fast != slow {
			t.Errorf("mode=%d: 빠른 경로와 폴백 경로 출력이 다름", mode)
		}
	}
}

// --- rgb2ansi256 팔레트 범위 테스트 ---

func TestRgb2Ansi256_Range(t *testing.T) {
	cases := []struct{ r, g, b uint8 }{
		{0, 0, 0},
		{255, 255, 255},
		{128, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{100, 200, 50},
	}
	for _, c := range cases {
		idx := rgb2ansi256(c.r, c.g, c.b)
		if idx < 16 || idx > 231 {
			t.Errorf("rgb2ansi256(%d,%d,%d) = %d, 범위 16-231 벗어남", c.r, c.g, c.b, idx)
		}
	}
}

// --- grayToAnsi256 경계값 테스트 ---

func TestGrayToAnsi256_Black(t *testing.T) {
	// luma < 8 → 16 (검정)
	if got := grayToAnsi256(0); got != 16 {
		t.Errorf("luma=0: got %d, want 16", got)
	}
	if got := grayToAnsi256(7); got != 16 {
		t.Errorf("luma=7: got %d, want 16", got)
	}
}

func TestGrayToAnsi256_White(t *testing.T) {
	// luma > 248 → 231 (흰색)
	if got := grayToAnsi256(249); got != 231 {
		t.Errorf("luma=249: got %d, want 231", got)
	}
	if got := grayToAnsi256(255); got != 231 {
		t.Errorf("luma=255: got %d, want 231", got)
	}
}

func TestGrayToAnsi256_MidRange(t *testing.T) {
	// 중간값 8~247은 232-255 범위 내여야 함 (248 이상은 흰색 231)
	for luma := uint8(8); luma <= 247; luma++ {
		got := grayToAnsi256(luma)
		if got < 232 || got > 255 {
			t.Errorf("luma=%d: got %d, 범위 232-255 벗어남", luma, got)
		}
	}
}

// --- toRGB8 luma 계산 테스트 ---

func TestToRGB8_PureColors(t *testing.T) {
	// 순수 빨강: luma ≈ 76 (299/1000 * 255)
	_, _, _, luma := toRGB8(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if luma < 70 || luma > 82 {
		t.Errorf("순수 빨강 luma: got %d, want ~76", luma)
	}
	// 흰색: luma = 255
	_, _, _, luma = toRGB8(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if luma != 255 {
		t.Errorf("흰색 luma: got %d, want 255", luma)
	}
	// 검정: luma = 0
	_, _, _, luma = toRGB8(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	if luma != 0 {
		t.Errorf("검정 luma: got %d, want 0", luma)
	}
}
