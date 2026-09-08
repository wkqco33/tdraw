package pnm

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

// decodeBytes는 바이트 슬라이스를 image.Decode 경로(등록된 포맷)로 디코딩한다.
func decodeBytes(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("디코딩 실패: %v", err)
	}
	if format == "" {
		t.Fatal("포맷이 감지되지 않음")
	}
	return img
}

// --- PGM (P2/P5) ---

func TestDecode_P2_ASCII(t *testing.T) {
	// 2x2 회색조: 0, 85, 170, 255
	data := []byte("P2\n# 주석 줄\n2 2\n255\n0 85\n170 255\n")
	img := decodeBytes(t, data)

	b := img.Bounds()
	if b.Dx() != 2 || b.Dy() != 2 {
		t.Fatalf("크기: got %dx%d, want 2x2", b.Dx(), b.Dy())
	}
	want := []uint8{0, 85, 170, 255}
	for i, v := range want {
		x, y := i%2, i/2
		if got := img.At(x, y).(color.Gray).Y; got != v {
			t.Errorf("픽셀(%d,%d): got %d, want %d", x, y, got, v)
		}
	}
}

func TestDecode_P5_Binary(t *testing.T) {
	// 2x2 회색조 바이너리
	data := []byte("P5\n2 2\n255\n\x00\x55\xaa\xff")
	img := decodeBytes(t, data)

	want := []uint8{0, 85, 170, 255}
	for i, v := range want {
		x, y := i%2, i/2
		if got := img.At(x, y).(color.Gray).Y; got != v {
			t.Errorf("픽셀(%d,%d): got %d, want %d", x, y, got, v)
		}
	}
}

func TestDecode_P5_16Bit(t *testing.T) {
	// maxval 65535, 2바이트 빅엔디언 샘플: 0, 32768, 65535
	// 32768*255/65535 = 127.5 → 127 (정수 나눗셈)
	data := []byte("P5\n1 3\n65535\n\x00\x00\x80\x00\xff\xff")
	img := decodeBytes(t, data)

	want := []uint8{0, 127, 255}
	for y, v := range want {
		if got := img.At(0, y).(color.Gray).Y; got != v {
			t.Errorf("픽셀(0,%d): got %d, want %d", y, got, v)
		}
	}
}

func TestDecode_P5_MaxvalScaling(t *testing.T) {
	// maxval 100: 값 50 → 127 (50*255/100 = 127.5 → 127)
	data := []byte("P5\n1 1\n100\n\x32")
	img := decodeBytes(t, data)
	if got := img.At(0, 0).(color.Gray).Y; got != 127 {
		t.Errorf("스케일링: got %d, want 127", got)
	}
}

func TestDecode_P5_ExtraWhitespace(t *testing.T) {
	// 헤더 뒤에 여러 공백이 있어도 정상 디코딩
	data := []byte("P5\n1 1\n255\n  \n\x7f")
	img := decodeBytes(t, data)
	if got := img.At(0, 0).(color.Gray).Y; got != 127 {
		t.Errorf("픽셀: got %d, want 127", got)
	}
}

// --- PPM (P3/P6) ---

func TestDecode_P3_ASCII(t *testing.T) {
	// 1x1 빨강
	data := []byte("P3\n1 1\n255\n255 0 0\n")
	img := decodeBytes(t, data)

	c := img.At(0, 0).(color.RGBA)
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("색: got %v, want {255 0 0 255}", c)
	}
}

func TestDecode_P6_Binary(t *testing.T) {
	// 1x2: 빨강, 초록
	data := []byte("P6\n1 2\n255\n\xff\x00\x00\x00\xff\x00")
	img := decodeBytes(t, data)

	want := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}}
	for y, wc := range want {
		if got := img.At(0, y).(color.RGBA); got != wc {
			t.Errorf("픽셀(0,%d): got %v, want %v", y, got, wc)
		}
	}
}

// --- PBM (P1/P4) ---

func TestDecode_P1_ASCII(t *testing.T) {
	// 2x2: 검정, 흰색, 흰색, 검정
	data := []byte("P1\n2 2\n0 1\n1 0\n")
	img := decodeBytes(t, data)

	want := []uint8{0, 255, 255, 0}
	for i, v := range want {
		x, y := i%2, i/2
		if got := img.At(x, y).(color.Gray).Y; got != v {
			t.Errorf("픽셀(%d,%d): got %d, want %d", x, y, got, v)
		}
	}
}

func TestDecode_P4_Binary(t *testing.T) {
	// 9x1: 1 0 1 0 1 0 1 0 1 → 바이트 0b10101010, 0b10000000
	data := []byte("P4\n9 1\n\xaa\x80")
	img := decodeBytes(t, data)

	want := []uint8{255, 0, 255, 0, 255, 0, 255, 0, 255}
	for x, v := range want {
		if got := img.At(x, 0).(color.Gray).Y; got != v {
			t.Errorf("픽셀(%d,0): got %d, want %d", x, got, v)
		}
	}
}

// --- DecodeConfig ---

func TestDecodeConfig(t *testing.T) {
	cases := []struct {
		name  string
		data  []byte
		model color.Model
	}{
		{"P2", []byte("P2\n4 3\n255\n"), color.GrayModel},
		{"P5", []byte("P5\n4 3\n255\n"), color.GrayModel},
		{"P3", []byte("P3\n4 3\n255\n"), color.RGBAModel},
		{"P6", []byte("P6\n4 3\n255\n"), color.RGBAModel},
		{"P1", []byte("P1\n4 3\n"), color.GrayModel},
		{"P4", []byte("P4\n4 3\n"), color.GrayModel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _, err := image.DecodeConfig(bytes.NewReader(tc.data))
			if err != nil {
				t.Fatalf("DecodeConfig 실패: %v", err)
			}
			if cfg.Width != 4 || cfg.Height != 3 {
				t.Errorf("크기: got %dx%d, want 4x3", cfg.Width, cfg.Height)
			}
			if cfg.ColorModel != tc.model {
				t.Errorf("컬러 모델: got %v, want %v", cfg.ColorModel, tc.model)
			}
		})
	}
}

// --- 오류 케이스 ---

func TestDecode_InvalidMagic(t *testing.T) {
	_, _, err := image.Decode(bytes.NewReader([]byte("X5\n1 1\n255\n\x00")))
	if err == nil {
		t.Error("잘못된 매직 넘버에 대해 에러가 발생해야 함")
	}
}

func TestDecode_Truncated(t *testing.T) {
	// 헤더는 유효하지만 데이터가 부족
	_, _, err := image.Decode(bytes.NewReader([]byte("P5\n2 2\n255\n\x00\x01")))
	if err == nil {
		t.Error("잘린 데이터에 대해 에러가 발생해야 함")
	}
}

func TestDecode_InvalidMaxval(t *testing.T) {
	_, _, err := image.Decode(bytes.NewReader([]byte("P5\n1 1\n0\n\x00")))
	if err == nil {
		t.Error("잘못된 maxval에 대해 에러가 발생해야 함")
	}
}

func TestDecode_ZeroDimension(t *testing.T) {
	_, _, err := image.Decode(bytes.NewReader([]byte("P5\n0 1\n255\n")))
	if err == nil {
		t.Error("0 크기에 대해 에러가 발생해야 함")
	}
}

func TestDecode_NonNumericHeader(t *testing.T) {
	_, _, err := image.Decode(bytes.NewReader([]byte("P5\nabc 1\n255\n")))
	if err == nil {
		t.Error("숫자가 아닌 헤더에 대해 에러가 발생해야 함")
	}
}
