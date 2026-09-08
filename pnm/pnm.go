// Package pnm은 Netpbm PNM 계열 이미지(PBM/PGM/PPM) 디코더를 제공한다.
// P1-P6 모든 변형(ASCII/바이너리)을 지원하며, image.RegisterFormat으로
// 등록되어 image.Decode를 통해 자동으로 사용된다.
//
//	P1/P4: PBM (흑백)  → image.Gray
//	P2/P5: PGM (회색조) → image.Gray
//	P3/P6: PPM (컬러)   → image.RGBA
package pnm

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
)

func init() {
	image.RegisterFormat("pbm", "P1", Decode, DecodeConfig)
	image.RegisterFormat("pgm", "P2", Decode, DecodeConfig)
	image.RegisterFormat("ppm", "P3", Decode, DecodeConfig)
	image.RegisterFormat("pbm", "P4", Decode, DecodeConfig)
	image.RegisterFormat("pgm", "P5", Decode, DecodeConfig)
	image.RegisterFormat("ppm", "P6", Decode, DecodeConfig)
}

// maxPixels는 메모리 보호를 위한 최대 픽셀 수다 (약 2.68억 픽셀).
const maxPixels = 1 << 28

// Decode는 PNM 이미지를 읽어 image.Image로 디코딩한다.
func Decode(r io.Reader) (image.Image, error) {
	br := bufio.NewReader(r)
	magic, w, h, maxval, err := decodeHeader(br)
	if err != nil {
		return nil, err
	}

	switch magic {
	case '1':
		return decodeP1(br, w, h)
	case '2':
		return decodeP2(br, w, h, maxval)
	case '3':
		return decodeP3(br, w, h, maxval)
	case '4':
		return decodeP4(br, w, h)
	case '5':
		return decodeP5(br, w, h, maxval)
	case '6':
		return decodeP6(br, w, h, maxval)
	}
	return nil, fmt.Errorf("pnm: 지원하지 않는 매직 넘버 P%c", magic)
}

// DecodeConfig는 PNM 이미지의 설정(크기, 컬러 모델)만 읽는다.
func DecodeConfig(r io.Reader) (image.Config, error) {
	br := bufio.NewReader(r)
	magic, w, h, _, err := decodeHeader(br)
	if err != nil {
		return image.Config{}, err
	}
	cm := color.GrayModel
	if magic == '3' || magic == '6' {
		cm = color.RGBAModel
	}
	return image.Config{ColorModel: cm, Width: w, Height: h}, nil
}

// decodeHeader는 매직 넘버와 헤더(너비, 높이, maxval)를 읽는다.
// PBM(P1/P4)은 maxval이 없으므로 1로 간주한다.
func decodeHeader(br *bufio.Reader) (magic byte, w, h, maxval int, err error) {
	var m [2]byte
	if _, err = io.ReadFull(br, m[:]); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 헤더 읽기 실패: %w", err)
	}
	if m[0] != 'P' || m[1] < '1' || m[1] > '6' {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 잘못된 매직 넘버 %q", m[:])
	}
	magic = m[1]

	if w, err = readInt(br); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 너비 읽기 실패: %w", err)
	}
	if h, err = readInt(br); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 높이 읽기 실패: %w", err)
	}
	if w <= 0 || h <= 0 {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 잘못된 크기 %dx%d", w, h)
	}
	if int64(w)*int64(h) > maxPixels {
		return 0, 0, 0, 0, fmt.Errorf("pnm: 이미지가 너무 큽니다 %dx%d", w, h)
	}

	maxval = 1
	if magic != '1' && magic != '4' {
		if maxval, err = readInt(br); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("pnm: maxval 읽기 실패: %w", err)
		}
		if maxval < 1 || maxval > 65535 {
			return 0, 0, 0, 0, fmt.Errorf("pnm: 잘못된 maxval %d", maxval)
		}
	}
	return magic, w, h, maxval, nil
}

// nextToken은 공백으로 구분된 다음 토큰을 읽는다. '#'으로 시작하는 주석 줄은 건너뛴다.
func nextToken(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		c, err := br.ReadByte()
		if err != nil {
			if err == io.EOF && sb.Len() > 0 {
				return sb.String(), nil
			}
			return "", err
		}
		if c == '#' {
			// 주석: 줄 끝까지 건너뜀
			for {
				c, err := br.ReadByte()
				if err != nil {
					return "", err
				}
				if c == '\n' {
					break
				}
			}
			continue
		}
		if isSpace(c) {
			if sb.Len() > 0 {
				return sb.String(), nil
			}
			continue
		}
		sb.WriteByte(c)
	}
}

// readInt는 다음 토큰을 정수로 읽는다.
func readInt(br *bufio.Reader) (int, error) {
	tok, err := nextToken(br)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(tok)
	if err != nil {
		return 0, fmt.Errorf("잘못된 숫자 %q", tok)
	}
	return n, nil
}

// skipWhitespace는 바이너리 데이터 시작 전의 공백을 건너뛴다.
// 헤더의 마지막 토큰 뒤에는 정확히 하나의 공백이 오지만, 일부 파일은
// 여러 공백을 포함하므로 관대하게 처리한다.
func skipWhitespace(br *bufio.Reader) error {
	for {
		c, err := br.ReadByte()
		if err != nil {
			return err
		}
		if !isSpace(c) {
			return br.UnreadByte()
		}
	}
}

// readSample은 바이너리 형식의 샘플 하나를 읽는다.
// maxval < 256이면 1바이트, 그 이상이면 2바이트(빅엔디언)다.
func readSample(br *bufio.Reader, maxval int) (int, error) {
	if maxval < 256 {
		b, err := br.ReadByte()
		return int(b), err
	}
	var buf [2]byte
	if _, err := io.ReadFull(br, buf[:]); err != nil {
		return 0, err
	}
	return int(buf[0])<<8 | int(buf[1]), nil
}

// scale8은 [0, maxval] 범위의 샘플 값을 8비트(0-255)로 변환한다.
func scale8(v, maxval int) uint8 {
	if maxval == 255 {
		return uint8(v)
	}
	return uint8(uint32(v) * 255 / uint32(maxval))
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// --- PBM (흑백) ---

// decodeP1은 ASCII PBM을 디코딩한다. 값 0=검정, 1=흰색.
func decodeP1(br *bufio.Reader, w, h int) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w]
		for x := range row {
			v, err := readInt(br)
			if err != nil {
				return nil, fmt.Errorf("pnm: PBM 데이터 읽기 실패: %w", err)
			}
			if v == 1 {
				row[x] = 255
			}
		}
	}
	return img, nil
}

// decodeP4는 바이너리 PBM을 디코딩한다. 픽셀은 MSB 우선으로 비트 패킹되고,
// 행은 바이트 경계로 패딩된다. 비트 1=흰색, 0=검정.
func decodeP4(br *bufio.Reader, w, h int) (image.Image, error) {
	if err := skipWhitespace(br); err != nil {
		return nil, fmt.Errorf("pnm: PBM 데이터 읽기 실패: %w", err)
	}
	img := image.NewGray(image.Rect(0, 0, w, h))
	rowBytes := (w + 7) / 8
	buf := make([]byte, rowBytes)
	for y := 0; y < h; y++ {
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, fmt.Errorf("pnm: PBM 데이터 읽기 실패: %w", err)
		}
		row := img.Pix[y*img.Stride : y*img.Stride+w]
		for x := range row {
			if (buf[x/8]>>(7-uint(x%8)))&1 == 1 {
				row[x] = 255
			}
		}
	}
	return img, nil
}

// --- PGM (회색조) ---

// decodeP2는 ASCII PGM을 디코딩한다.
func decodeP2(br *bufio.Reader, w, h, maxval int) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w]
		for x := range row {
			v, err := readInt(br)
			if err != nil {
				return nil, fmt.Errorf("pnm: PGM 데이터 읽기 실패: %w", err)
			}
			row[x] = scale8(v, maxval)
		}
	}
	return img, nil
}

// decodeP5는 바이너리 PGM을 디코딩한다.
func decodeP5(br *bufio.Reader, w, h, maxval int) (image.Image, error) {
	if err := skipWhitespace(br); err != nil {
		return nil, fmt.Errorf("pnm: PGM 데이터 읽기 실패: %w", err)
	}
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w]
		for x := range row {
			v, err := readSample(br, maxval)
			if err != nil {
				return nil, fmt.Errorf("pnm: PGM 데이터 읽기 실패: %w", err)
			}
			row[x] = scale8(v, maxval)
		}
	}
	return img, nil
}

// --- PPM (컬러) ---

// decodeP3은 ASCII PPM을 디코딩한다.
func decodeP3(br *bufio.Reader, w, h, maxval int) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w*4]
		for x := 0; x < w; x++ {
			r, err := readInt(br)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			g, err := readInt(br)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			b, err := readInt(br)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			i := x * 4
			row[i] = scale8(r, maxval)
			row[i+1] = scale8(g, maxval)
			row[i+2] = scale8(b, maxval)
			row[i+3] = 255
		}
	}
	return img, nil
}

// decodeP6은 바이너리 PPM을 디코딩한다. 픽셀은 R,G,B 순서로 인터리브된다.
func decodeP6(br *bufio.Reader, w, h, maxval int) (image.Image, error) {
	if err := skipWhitespace(br); err != nil {
		return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+w*4]
		for x := 0; x < w; x++ {
			r, err := readSample(br, maxval)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			g, err := readSample(br, maxval)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			b, err := readSample(br, maxval)
			if err != nil {
				return nil, fmt.Errorf("pnm: PPM 데이터 읽기 실패: %w", err)
			}
			i := x * 4
			row[i] = scale8(r, maxval)
			row[i+1] = scale8(g, maxval)
			row[i+2] = scale8(b, maxval)
			row[i+3] = 255
		}
	}
	return img, nil
}
