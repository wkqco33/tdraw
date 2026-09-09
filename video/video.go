// Package video는 프레임 스트림을 터미널에 실시간 재생하는 엔진을 제공한다.
// 디코딩 백엔드는 FrameSource 인터페이스 뒤로 숨겨져 있으므로, 엔진 자체는
// 순수 Go로 유지되며 어떤 소스와도 조립될 수 있다.
package video

import (
	"context"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"time"

	"github.com/wkqco33/tdraw/render"
)

// Spec는 비디오 소스의 메타 정보다.
type Spec struct {
	Width  int
	Height int
	FPS    float64
}

// FrameSource는 재생 가능한 프레임 스트림이다.
type FrameSource interface {
	// Spec은 소스의 원본 해상도와 초당 프레임 수를 반환한다.
	Spec() Spec
	// Rotation은 메타데이터상 회전 각도(0/90/180/270, 시계 방향)를 반환한다.
	Rotation() int
	// Next는 다음 프레임을 RGB24로 반환한다. targetW/targetH가 0보다 크면
	// 소스는 해당 크기로 스케일한다(비율은 보장되지 않으므로 호출자가
	// TargetSize로 정한다). 스트림 끝이면 반드시 io.EOF를 반환한다.
	Next(targetW, targetH int) (rgb []byte, w, h, stride int, err error)
	// Rewind는 스트림의 처음으로 되돌린다.
	Rewind() error
	// Close는 소스가 잡은 자원을 해제한다.
	Close()
}

// Options는 재생 설정이다.
type Options struct {
	// Width는 출력 너비(셀 수)다. 0이면 터미널 너비를 쓴다.
	Width int
	// Mode는 렌더링 컬러 모드다.
	Mode render.ColorMode
	// Loop가 true면 스트림 끝에서 처음부터 다시 재생한다.
	Loop bool
	// TargetW/TargetH는 디코더에 요청할 프레임 크기(픽셀)다.
	// 0이면 터미널 크기와 소스 해상도로 자동 계산한다.
	TargetW, TargetH int
	// ReserveRows는 자동 계산 시 프레임 위에 남겨둘 행 수다(메타 박스 등).
	ReserveRows int
	// Out은 출력 대상이다. nil이면 os.Stdout를 쓴다.
	Out io.Writer
}

// TargetSize는 원본 (vw, vh)을 셀 예산 (maxCols, maxRows) 안에 비율을 유지하며
// 맞추기 위한 디코더 스케일 크기를 반환한다. maxRows는 문자 행 수이고,
// half-block 렌더링으로 한 행당 2픽셀을 표현할 수 있다.
//
// rot은 소스 메타데이터의 시계 방향 회전 각도(0/90/180/270)다. 회전 후 표시
// 높이가 짝수여야 렌더링할 수 있으므로, 90/270에서는 (회전 후 기준으로) 너비를
// 짝수로 맞춘 디코더 타깃을 반환한다. 반환값은 항상 2 이상이다.
func TargetSize(vw, vh, rot, maxCols, maxRows int) (tw, th int) {
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}
	if maxCols < 2 {
		maxCols = 2
	}
	if maxRows < 1 {
		maxRows = 1
	}
	// half-block: 문자 행 1개 = 픽셀 2줄
	pw, ph := maxCols, maxRows*2

	if rot%180 == 90 { // 90/270: 표시 시 가로세로가 바뀐다
		dw, dh := fit(vh, vw, pw, ph)
		// 회전 후 높이(dh)는 회전 전 너비(tw)에서 온다.
		return even(dh), dw
	}
	tw, th = fit(vw, vh, pw, ph)
	return tw, even(th)
}

// fit는 (aw, ah)를 (pw, ph) 안에 비율을 유지하며 맞춘 크기를 반환한다.
func fit(aw, ah, pw, ph int) (w, h int) {
	scale := math.Min(float64(pw)/float64(aw), float64(ph)/float64(ah))
	w = int(math.Round(float64(aw) * scale))
	h = int(math.Round(float64(ah) * scale))
	if w < 2 {
		w = 2
	}
	if h < 2 {
		h = 2
	}
	return w, h
}

// even은 n 이하의 가장 큰 짝수를 반환한다(최소 2).
func even(n int) int {
	if n < 2 {
		return 2
	}
	return n &^ 1
}

// defaultOut은 기본 출력 대상이다. 터미널 렌더링이므로 stdout에 직접 쓴다.
func defaultOut() io.Writer { return os.Stdout }

// Play는 src의 프레임을 out에 실시간 재생한다.
// opts.Loop가 true면 스트림 끝에서 되감아 계속 재생하고, false면 스트림 끝에서
// 정상 종료한다. ctx가 취소되면(Ctrl+C 등) 즉시 멈춘다.
func Play(ctx context.Context, src FrameSource, opts Options) error {
	out := opts.Out
	if out == nil {
		out = defaultOut()
	}

	spec := src.Spec()
	rot := ((src.Rotation() % 360) + 360) % 360

	tw, th := opts.TargetW, opts.TargetH
	if tw <= 0 || th <= 0 {
		cols, rows := render.TermSize()
		if opts.Width > 0 {
			cols = opts.Width
		}
		tw, th = TargetSize(spec.Width, spec.Height, rot, cols, rows-opts.ReserveRows)
	}

	fps := spec.FPS
	if fps <= 0 {
		fps = 30
	}
	frameDur := time.Duration(float64(time.Second) / fps)

	// 첫 프레임 크기에 맞춰 변환 버퍼를 준비한다. 스케일 타깃이 고정되어 있어
	// 프레임 크기는 재생 중 변하지 않는다.
	img := image.NewRGBA(image.Rect(0, 0, tw, th))

	fmt.Fprint(out, "\033[?25l") // 커서 숨김
	defer fmt.Fprint(out, "\033[?25h")

	first := true
	displayRows := 0
	for {
		frameStart := time.Now()

		rgb, w, h, stride, err := src.Next(tw, th)
		if err == io.EOF {
			if !opts.Loop {
				return nil
			}
			if err := src.Rewind(); err != nil {
				return fmt.Errorf("되감기 실패: %w", err)
			}
			continue
		}
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if w != img.Bounds().Dx() || h != img.Bounds().Dy() {
			img = image.NewRGBA(image.Rect(0, 0, w, h))
		}
		RGB24ToRGBA(rgb, w, stride, img)
		frameImg := image.Image(img)
		if rot != 0 {
			frameImg = rotateRGBA(img, rot)
		}

		s := render.RenderString(frameImg, opts.Mode)
		rows := frameImg.Bounds().Dy() / 2
		if first {
			displayRows = rows
		} else if _, err := fmt.Fprintf(out, "\033[%dA", displayRows); err != nil {
			return nil // 파이프 끊김 등 출력 실패 시 정상 종료
		}
		if _, err := fmt.Fprint(out, s); err != nil {
			return nil
		}
		first = false

		// fps 페이싱: 프레임 처리에 남은 시간만큼 잠들고, 취소를 감시한다.
		if d := frameDur - time.Since(frameStart); d > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d):
			}
		} else if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// RGB24ToRGBA는 RGB24 버퍼(stride 포함)를 img에 채운다.
// img 크기는 (w, 높이)와 일치해야 하며, 높이는 버퍼 크기에서 유도된다.
func RGB24ToRGBA(rgb []byte, w, stride int, img *image.RGBA) {
	b := img.Bounds()
	for y := range b.Dy() {
		row := rgb[y*stride:]
		dst := img.Pix[y*img.Stride:]
		for x := range w {
			i, o := x*3, x*4
			dst[o+0] = row[i+0]
			dst[o+1] = row[i+1]
			dst[o+2] = row[i+2]
			dst[o+3] = 255
		}
	}
}

// rotateRGBA는 img를 deg(시계 방향 90/180/270)만큼 회전한 사본을 반환한다.
// deg가 0이면 img 자체를 반환한다.
func rotateRGBA(img *image.RGBA, deg int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	switch deg % 360 {
	case 90:
		out := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := range h {
			for x := range w {
				// 시계 90도: (x, y) -> (h-1-y, x)
				copy(out.Pix[(x*out.Stride)+(h-1-y)*4:], img.Pix[(y*img.Stride)+(x*4):(y*img.Stride)+(x*4)+4])
			}
		}
		return out
	case 180:
		out := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := range h {
			for x := range w {
				// 시계 180도: (x, y) -> (w-1-x, h-1-y)
				copy(out.Pix[((h-1-y)*out.Stride)+((w-1-x)*4):], img.Pix[(y*img.Stride)+(x*4):(y*img.Stride)+(x*4)+4])
			}
		}
		return out
	case 270:
		out := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := range h {
			for x := range w {
				// 시계 270도: (x, y) -> (y, w-1-x)
				copy(out.Pix[((w-1-x)*out.Stride)+(y*4):], img.Pix[(y*img.Stride)+(x*4):(y*img.Stride)+(x*4)+4])
			}
		}
		return out
	default:
		return img
	}
}
