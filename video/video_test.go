package video

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"
	"time"
)

func TestTargetSize(t *testing.T) {
	tests := []struct {
		name             string
		vw, vh, rot      int
		maxCols, maxRows int
		wantW, wantH     int
	}{
		{"16:9 대략 종횡비 유지", 1920, 1080, 0, 80, 20, 71, 40},
		{"세로 영상은 폭이 줄어든다", 1080, 1920, 0, 80, 20, 23, 40},
		{"작은 영상은 예산에 맞춰 확대한다(imgutil.Resize와 동일)", 40, 20, 0, 80, 20, 80, 40},
		{"높이는 항상 짝수", 100, 101, 0, 200, 200, 200, 202},
		{"90도 회전 시 가로세로 교환", 1920, 1080, 90, 80, 20, 40, 23},
		{"270도 회전도 동일", 1920, 1080, 270, 80, 20, 40, 23},
		{"180도 회전은 교환 없음", 1920, 1080, 180, 80, 20, 71, 40},
		{"터미널이 극단적으로 작아도 최소 크기", 1920, 1080, 0, 1, 1, 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw, th := TargetSize(tt.vw, tt.vh, tt.rot, tt.maxCols, tt.maxRows)
			if tw != tt.wantW || th != tt.wantH {
				t.Fatalf("TargetSize(%d,%d,rot=%d,%d,%d) = %dx%d, want %dx%d",
					tt.vw, tt.vh, tt.rot, tt.maxCols, tt.maxRows, tw, th, tt.wantW, tt.wantH)
			}

			// 표시(회전 후) 높이는 렌더링을 위해 항상 짝수여야 한다.
			displayH := th
			if tt.rot%180 == 90 {
				displayH = tw
			}
			if displayH%2 != 0 {
				t.Fatalf("표시 높이 %d가 짝수가 아니다", displayH)
			}

			// 결과는 셀 예산 안에 들어와야 한다(최소 클램프 반영).
			effCols, effRows := tt.maxCols, tt.maxRows
			if effCols < 2 {
				effCols = 2
			}
			if effRows < 1 {
				effRows = 1
			}
			dw, dh := tw, th
			if tt.rot%180 == 90 {
				dw, dh = th, tw
			}
			if dw > effCols || dh > effRows*2 {
				t.Fatalf("결과 %dx%d(회전 후 %dx%d)가 예산 %dx%d을 초과한다",
					tw, th, dw, dh, effCols, effRows*2)
			}
		})
	}
}

func TestRGB24ToRGBA(t *testing.T) {
	// 2x2 프레임, stride 7(패딩 1바이트)로 패딩을 무시하는지 검사한다.
	rgb := []byte{
		10, 20, 30, 40, 50, 60, 0xFF, // 행 0 + 패딩
		70, 80, 90, 100, 110, 120, 0xFF, // 행 1 + 패딩
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	RGB24ToRGBA(rgb, 2, 7, img)

	want := [][4]uint8{
		{10, 20, 30, 255},
		{40, 50, 60, 255},
		{70, 80, 90, 255},
		{100, 110, 120, 255},
	}
	for i, w := range want {
		x, y := i%2, i/2
		r, g, b, a := img.At(x, y).RGBA()
		if uint8(r>>8) != w[0] || uint8(g>>8) != w[1] || uint8(b>>8) != w[2] || uint8(a>>8) != w[3] {
			t.Fatalf("pixel(%d,%d) = (%d,%d,%d,%d), want %v", x, y, r>>8, g>>8, b>>8, a>>8, w)
		}
	}
}

func TestRotateRGBA(t *testing.T) {
	// 2x3 이미지: 각 픽셀 R 값을 좌표로 식별한다(왼쪽 위 = 1).
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	vals := [][]uint8{
		{1, 2},
		{3, 4},
		{5, 6},
	}
	for y := range vals {
		for x := range vals[y] {
			img.SetRGBA(x, y, color.RGBA{R: vals[y][x], A: 255})
		}
	}

	red := func(im image.Image, x, y int) uint8 {
		r, _, _, _ := im.At(x, y).RGBA()
		return uint8(r >> 8)
	}

	// 시계 90도: 2x3 -> 3x2, (x,y) -> (h-1-y, x)
	r90 := rotateRGBA(img, 90)
	if r90.Bounds().Dx() != 3 || r90.Bounds().Dy() != 2 {
		t.Fatalf("90도 회전 크기 = %v", r90.Bounds())
	}
	for _, c := range []struct{ x, y, want int }{
		{0, 0, 5}, {0, 1, 6}, {1, 0, 3}, {1, 1, 4}, {2, 0, 1}, {2, 1, 2},
	} {
		if got := red(r90, c.x, c.y); got != uint8(c.want) {
			t.Fatalf("90도 (%d,%d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}

	// 시계 180도: (x,y) -> (w-1-x, h-1-y)
	r180 := rotateRGBA(img, 180)
	for _, c := range []struct{ x, y, want int }{
		{1, 2, 1}, {0, 2, 2}, {1, 1, 3}, {0, 1, 4}, {1, 0, 5}, {0, 0, 6},
	} {
		if got := red(r180, c.x, c.y); got != uint8(c.want) {
			t.Fatalf("180도 (%d,%d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}

	// 시계 270도: 3x2, (x,y) -> (y, w-1-x)
	r270 := rotateRGBA(img, 270)
	if r270.Bounds().Dx() != 3 || r270.Bounds().Dy() != 2 {
		t.Fatalf("270도 회전 크기 = %v", r270.Bounds())
	}
	for _, c := range []struct{ x, y, want int }{
		{0, 1, 1}, {0, 0, 2}, {1, 1, 3}, {1, 0, 4}, {2, 1, 5}, {2, 0, 6},
	} {
		if got := red(r270, c.x, c.y); got != uint8(c.want) {
			t.Fatalf("270도 (%d,%d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}

	// 0도는 원본 그대로
	if rotateRGBA(img, 0) != image.Image(img) {
		t.Fatal("0도 회전은 원본을 반환해야 한다")
	}
}

// fakeSource는 테스트용 FrameSource다.
type fakeSource struct {
	spec         Spec
	frames       int
	next         int
	nexts        int
	rewinds      int
	failRewindAt int // 이 번호(1부터)의 Rewind에서 에러를 반환한다
}

func newFakeSource(frames int) *fakeSource {
	// 페이싱 검사와 분리된 테스트에서는 빠르게 돌도록 fps를 높게 잡는다.
	return &fakeSource{frames: frames, spec: Spec{Width: 4, Height: 4, FPS: 1000}}
}

func (f *fakeSource) Spec() Spec { return f.spec }

func (f *fakeSource) Rotation() int { return 0 }

func (f *fakeSource) Next(tw, th int) ([]byte, int, int, int, error) {
	f.nexts++
	if f.next >= f.frames {
		return nil, 0, 0, 0, io.EOF
	}
	f.next++
	return make([]byte, 4*4*3), 4, 4, 4 * 3, nil
}

func (f *fakeSource) Rewind() error {
	f.rewinds++
	if f.failRewindAt > 0 && f.rewinds >= f.failRewindAt {
		return errors.New("rewind 실패")
	}
	f.next = 0
	return nil
}

func (f *fakeSource) Close() {}

func TestPlayUntilEOF(t *testing.T) {
	src := newFakeSource(3)
	var out bytes.Buffer

	err := Play(context.Background(), src, Options{Out: &out})
	if err != nil {
		t.Fatalf("Play 실패: %v", err)
	}
	if src.nexts != 4 { // 3프레임 + EOF 1회
		t.Fatalf("Next 호출 수 = %d, want 4", src.nexts)
	}
	if !strings.Contains(out.String(), "▀") {
		t.Fatal("출력에 half-block 문자가 없다")
	}
}

func TestPlayLoopRewinds(t *testing.T) {
	// 2프레임짜리를 세 바퀴 재생하고, 세 번째 Rewind에서 실패시켜 종료를
	// 결정적으로 만든다.
	src := newFakeSource(2)
	src.failRewindAt = 3
	var out bytes.Buffer

	err := Play(context.Background(), src, Options{Out: &out, Loop: true})
	if err == nil || !strings.Contains(err.Error(), "되감기") {
		t.Fatalf("세 번째 되감기에서 오류를 기대했으나: %v", err)
	}
	if src.nexts != 9 { // 라운드별 2프레임+EOF = 3, 세 라운드
		t.Fatalf("Next 호출 수 = %d, want 9", src.nexts)
	}
	if src.rewinds != 3 {
		t.Fatalf("Rewind 호출 수 = %d, want 3", src.rewinds)
	}
}

func TestPlayCancelBeforeFrame(t *testing.T) {
	src := newFakeSource(5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 시작 전에 취소

	var out bytes.Buffer
	err := Play(ctx, src, Options{Out: &out})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context.Canceled를 기대했으나: %v", err)
	}
	if src.nexts != 1 { // 첫 프레임 읽은 뒤 취소를 확인한다
		t.Fatalf("Next 호출 수 = %d, want 1", src.nexts)
	}
}

func TestPlayPacesToFPS(t *testing.T) {
	src := newFakeSource(2)
	// fps를 20으로 낮춰 프레임당 50ms가 걸리는지 확인한다.
	src.spec = Spec{Width: 4, Height: 4, FPS: 20}

	var out bytes.Buffer
	start := time.Now()
	if err := Play(context.Background(), src, Options{Out: &out}); err != nil {
		t.Fatalf("Play 실패: %v", err)
	}
	// 프레임 2개 × 50ms = 100ms. 여유를 두고 검사한다.
	if elapsed := time.Since(start); elapsed < 90*time.Millisecond {
		t.Fatalf("너무 빨리 끝났다: %v (페이싱이 동작하지 않는다)", elapsed)
	}
}
