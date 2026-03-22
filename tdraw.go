// Package tdraw는 이미지를 터미널에 렌더링하는 통합 API를 제공한다.
package tdraw

import (
	"image"
	"io"

	"github.com/wkqco/tdraw/imgutil"
	"github.com/wkqco/tdraw/render"
)

type ColorMode = render.ColorMode

const (
	TrueColor = render.ColorTruecolor
	Color256  = render.Color256
	Gray      = render.ColorGray
)

type Options struct {
	Width     int       // 0이면 자동(터미널 폭)
	Height    int       // 0이면 자동(비율 유지, Width*4)
	ColorMode ColorMode // 기본값: TrueColor
}

// DrawFile은 파일 경로의 이미지를 w에 렌더링한다.
func DrawFile(w io.Writer, path string, opts Options) error {
	info, err := imgutil.Load(path)
	if err != nil {
		return err
	}
	Draw(w, info.Image, opts)
	return nil
}

// Draw는 image.Image를 w에 렌더링한다.
func Draw(w io.Writer, img image.Image, opts Options) {
	targetW := opts.Width
	targetH := opts.Height

	if targetW <= 0 {
		cols, _ := render.TermSize()
		targetW = cols
	}
	if targetH <= 0 {
		targetH = targetW * 4
	}

	resized := imgutil.Resize(img, targetW, targetH)
	render.Render(w, resized, opts.ColorMode)
}

// TermSize는 현재 터미널 크기를 반환한다.
func TermSize() (cols, rows int) {
	return render.TermSize()
}
