// Package tdraw는 이미지를 터미널에 렌더링하는 통합 API를 제공한다.
package tdraw

import (
	"context"
	"image"
	"io"

	"github.com/wkqco33/tdraw/imgutil"
	"github.com/wkqco33/tdraw/render"
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

// PlayGIFFile은 GIF 파일을 w에 애니메이션으로 무한 반복 재생한다.
// ctx가 취소되면 재생을 멈춘다.
func PlayGIFFile(ctx context.Context, w io.Writer, path string, opts Options) error {
	anim, err := imgutil.LoadGIF(path)
	if err != nil {
		return err
	}

	targetW := opts.Width
	targetH := opts.Height
	if targetW <= 0 {
		cols, _ := render.TermSize()
		targetW = cols
	}
	if targetH <= 0 {
		targetH = targetW * 4
	}

	frames := make([]image.Image, len(anim.Frames))
	for i, f := range anim.Frames {
		frames[i] = imgutil.Resize(f, targetW, targetH)
	}

	render.PlayGIF(ctx, w, frames, anim.Delays, opts.ColorMode)
	return nil
}

// TermSize는 현재 터미널 크기를 반환한다.
func TermSize() (cols, rows int) {
	return render.TermSize()
}
