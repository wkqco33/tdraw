package render

import (
	"image"
	"image/color"
	"testing"
)

// newGradientImage는 벤치마크용 그라데이션 RGBA 이미지를 반환한다.
// 단색이 아니라 픽셀마다 색이 달라 색 시퀀스 생성 비용을 제대로 측정한다.
func newGradientImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x),
				G: uint8(y),
				B: uint8(x ^ y),
				A: 255,
			})
		}
	}
	return img
}

func benchmarkRenderString(b *testing.B, mode ColorMode) {
	// 80열 × 80행(=160픽셀 높이) 대표 크기
	img := newGradientImage(80, 160)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderString(img, mode)
	}
}

func BenchmarkRenderString_Truecolor(b *testing.B) { benchmarkRenderString(b, ColorTruecolor) }
func BenchmarkRenderString_Color256(b *testing.B)  { benchmarkRenderString(b, Color256) }
func BenchmarkRenderString_Gray(b *testing.B)      { benchmarkRenderString(b, ColorGray) }
