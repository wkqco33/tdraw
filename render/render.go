package render

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type ColorMode int

const (
	ColorTruecolor ColorMode = iota
	Color256
	ColorGray
)

// TermSize는 현재 터미널 크기를 반환한다. 실패 시 기본값 80x24.
func TermSize() (cols, rows int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}

// Render는 img를 터미널에 half-block 방식으로 출력한다.
// img의 높이는 반드시 짝수여야 한다.
func Render(w io.Writer, img image.Image, mode ColorMode) {
	fmt.Fprint(w, RenderString(img, mode))
}

// RenderString은 img를 half-block 방식으로 변환한 문자열을 반환한다.
// 행 수만큼 \n으로 끝나며, 애니메이션 재생 시 프레임을 미리 캐싱하는 데 쓴다.
//
// 핫루프이므로 픽셀당 fmt.Sprintf/인터페이스 할당을 피한다. img가 *image.RGBA이면
// RGBAAt로 직접 접근하고(빠른 경로), ANSI 시퀀스는 strings.Builder에 바로 쓴다.
func RenderString(img image.Image, mode ColorMode) string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var sb strings.Builder
	// 픽셀당 색 시퀀스(최대 ~40B) + 반블록 문자, 행마다 리셋. 대략치로 사전 할당.
	rows := height / 2
	sb.Grow(rows * (width*44 + 8))

	rgba, fast := img.(*image.RGBA)

	for y := bounds.Min.Y; y < bounds.Min.Y+height-1; y += 2 {
		for x := bounds.Min.X; x < bounds.Min.X+width; x++ {
			var ur, ug, ub, lr, lg, lb uint8
			if fast {
				u := rgba.RGBAAt(x, y)
				l := rgba.RGBAAt(x, y+1)
				ur, ug, ub = u.R, u.G, u.B
				lr, lg, lb = l.R, l.G, l.B
			} else {
				ur, ug, ub, _ = toRGB8(img.At(x, y))
				lr, lg, lb, _ = toRGB8(img.At(x, y+1))
			}

			switch mode {
			case ColorTruecolor:
				writeTruecolor(&sb, ur, ug, ub, lr, lg, lb)
			case Color256:
				writeColor256(&sb, rgb2ansi256(ur, ug, ub), rgb2ansi256(lr, lg, lb))
			case ColorGray:
				writeColor256(&sb, grayToAnsi256(luma(ur, ug, ub)), grayToAnsi256(luma(lr, lg, lb)))
			}
			sb.WriteString("▀")
		}
		sb.WriteString("\033[0m\n")
	}

	return sb.String()
}

// PlayGIF는 합성된 프레임들을 커서 이동 기반으로 무한 반복 재생한다.
// 프레임은 모두 동일한 크기여야 하며, ctx가 취소되면(Ctrl+C 등) 재생을 멈춘다.
func PlayGIF(ctx context.Context, w io.Writer, frames []image.Image, delays []time.Duration, mode ColorMode) {
	if len(frames) == 0 {
		return
	}

	// 프레임을 미리 문자열로 렌더링해 재생 중 CPU 부담을 줄인다.
	strs := make([]string, len(frames))
	for i, f := range frames {
		strs[i] = RenderString(f, mode)
	}
	// 프레임 높이(터미널 행 수): half-block이므로 픽셀 높이/2
	rows := frames[0].Bounds().Dy() / 2

	if _, err := fmt.Fprint(w, "\033[?25l"); err != nil { // 커서 숨김
		return
	}
	defer fmt.Fprint(w, "\033[?25h") // 종료 시 커서 복원

	first := true
	for {
		for i, s := range strs {
			if !first {
				// 커서를 프레임 맨 위로. Write 실패(파이프 끊김 등) 시 정상 종료.
				if _, err := fmt.Fprintf(w, "\033[%dA", rows); err != nil {
					return
				}
			}
			first = false
			if _, err := fmt.Fprint(w, s); err != nil {
				return
			}

			// delays가 frames보다 짧을 수 있으므로 방어(부족분 기본 100ms)
			d := 100 * time.Millisecond
			if i < len(delays) {
				d = delays[i]
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(d):
			}
		}
	}
}

// writeTruecolor는 상/하 픽셀의 24-bit RGB ANSI 시퀀스를 sb에 직접 쓴다.
// (전경=상단, 배경=하단; 반블록 ▀ 기준)
func writeTruecolor(sb *strings.Builder, fr, fg, fb, br, bg, bb uint8) {
	sb.WriteString("\033[38;2;")
	writeDec(sb, fr)
	sb.WriteByte(';')
	writeDec(sb, fg)
	sb.WriteByte(';')
	writeDec(sb, fb)
	sb.WriteString("m\033[48;2;")
	writeDec(sb, br)
	sb.WriteByte(';')
	writeDec(sb, bg)
	sb.WriteByte(';')
	writeDec(sb, bb)
	sb.WriteByte('m')
}

// writeColor256은 상/하 픽셀의 ANSI 256색 인덱스 시퀀스를 sb에 직접 쓴다.
func writeColor256(sb *strings.Builder, fg, bg int) {
	sb.WriteString("\033[38;5;")
	writeDec(sb, uint8(fg))
	sb.WriteString("m\033[48;5;")
	writeDec(sb, uint8(bg))
	sb.WriteByte('m')
}

// writeDec는 0~255 정수를 10진 ASCII로 sb에 append한다(할당 없음).
func writeDec(sb *strings.Builder, n uint8) {
	switch {
	case n >= 100:
		sb.WriteByte('0' + n/100)
		sb.WriteByte('0' + (n/10)%10)
		sb.WriteByte('0' + n%10)
	case n >= 10:
		sb.WriteByte('0' + n/10)
		sb.WriteByte('0' + n%10)
	default:
		sb.WriteByte('0' + n)
	}
}

// luma는 RGB의 ITU-R BT.601 밝기값을 반환한다.
func luma(r, g, b uint8) uint8 {
	return uint8((uint32(r)*299 + uint32(g)*587 + uint32(b)*114) / 1000)
}

// toRGB8은 임의의 color.Color에서 8-bit RGB와 luma를 추출한다(폴백 경로/테스트용).
func toRGB8(c color.Color) (r, g, b, lum uint8) {
	r16, g16, b16, _ := c.RGBA()
	r8 := uint8(r16 >> 8)
	g8 := uint8(g16 >> 8)
	b8 := uint8(b16 >> 8)
	return r8, g8, b8, luma(r8, g8, b8)
}

// rgb2ansi256은 RGB를 ANSI 256색 팔레트 인덱스로 변환한다.
func rgb2ansi256(r, g, b uint8) int {
	// 6x6x6 컬러 큐브 (16-231)
	ri := int(r) * 5 / 255
	gi := int(g) * 5 / 255
	bi := int(b) * 5 / 255
	return 16 + 36*ri + 6*gi + bi
}

// grayToAnsi256은 밝기값을 ANSI 256 회색조 범위(232-255)로 변환한다.
func grayToAnsi256(luma uint8) int {
	if luma < 8 {
		return 16 // 검정
	}
	if luma >= 248 {
		return 231 // 흰색
	}
	return 232 + int(luma-8)/10
}
