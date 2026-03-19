package render

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"strings"

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
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var sb strings.Builder

	for y := bounds.Min.Y; y < bounds.Min.Y+height-1; y += 2 {
		for x := bounds.Min.X; x < bounds.Min.X+width; x++ {
			upper := img.At(x, y)
			lower := img.At(x, y+1)

			switch mode {
			case ColorTruecolor:
				sb.WriteString(truecolorSeq(upper, lower))
			case Color256:
				sb.WriteString(color256Seq(upper, lower))
			case ColorGray:
				sb.WriteString(graySeq(upper, lower))
			}
			sb.WriteString("▀")
		}
		sb.WriteString("\033[0m\n")
	}

	fmt.Fprint(w, sb.String())
}

func truecolorSeq(fg, bg color.Color) string {
	fr, fg_, fb, _ := toRGB8(fg)
	br, bg_, bb, _ := toRGB8(bg)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm", fr, fg_, fb, br, bg_, bb)
}

func color256Seq(fg, bg color.Color) string {
	fr, fg_, fb, _ := toRGB8(fg)
	br, bg_, bb, _ := toRGB8(bg)
	return fmt.Sprintf("\033[38;5;%dm\033[48;5;%dm", rgb2ansi256(fr, fg_, fb), rgb2ansi256(br, bg_, bb))
}

func graySeq(fg, bg color.Color) string {
	_, _, _, fa := toRGB8(fg)
	_, _, _, ba := toRGB8(bg)
	fg256 := grayToAnsi256(fa)
	bg256 := grayToAnsi256(ba)
	return fmt.Sprintf("\033[38;5;%dm\033[48;5;%dm", fg256, bg256)
}

func toRGB8(c color.Color) (r, g, b, luma uint8) {
	r16, g16, b16, _ := c.RGBA()
	r8 := uint8(r16 >> 8)
	g8 := uint8(g16 >> 8)
	b8 := uint8(b16 >> 8)
	// ITU-R BT.601 luma
	l := uint8((uint32(r8)*299 + uint32(g8)*587 + uint32(b8)*114) / 1000)
	return r8, g8, b8, l
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
