package imgutil

import (
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	_ "github.com/wkqco33/tdraw/pnm" // PBM/PGM/PPM (P1-P6)
)

type ImageInfo struct {
	Path   string
	Format string
	Width  int
	Height int
	Image  image.Image
}

func Load(path string) (*ImageInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))

	var img image.Image
	var format string

	if ext == ".gif" {
		gifs, err := gif.DecodeAll(f)
		if err != nil {
			return nil, fmt.Errorf("GIF 디코딩 실패: %w", err)
		}
		img = gifs.Image[0]
		format = "gif"
	} else {
		img, format, err = image.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("이미지 디코딩 실패: %w", err)
		}
	}

	bounds := img.Bounds()
	return &ImageInfo{
		Path:   path,
		Format: format,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Image:  img,
	}, nil
}

// GIFAnim은 합성 완료된 GIF 애니메이션 프레임들을 담는다.
type GIFAnim struct {
	Frames []image.Image   // 각 프레임을 전체 캔버스 기준으로 합성한 완성본
	Delays []time.Duration // 프레임별 표시 시간
	Width  int
	Height int
}

// LoadGIF는 GIF 파일의 모든 프레임을 disposal method에 따라 합성하여 반환한다.
// 각 프레임은 image.Paletted의 부분 영역만 담고 있으므로 전체 캔버스에 누적
// 합성한 스냅샷을 만들어야 실제로 보이는 화면이 된다.
func LoadGIF(path string) (*GIFAnim, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, fmt.Errorf("GIF 디코딩 실패: %w", err)
	}
	if len(g.Image) == 0 {
		return nil, fmt.Errorf("GIF에 프레임이 없습니다")
	}

	w, h := g.Config.Width, g.Config.Height
	canvas := image.NewRGBA(image.Rect(0, 0, w, h))

	anim := &GIFAnim{
		Frames: make([]image.Image, 0, len(g.Image)),
		Delays: make([]time.Duration, 0, len(g.Image)),
		Width:  w,
		Height: h,
	}

	for i, src := range g.Image {
		// DisposalPrevious 대비: 그리기 전 캔버스 상태 백업
		var backup *image.RGBA
		if g.Disposal[i] == gif.DisposalPrevious {
			backup = image.NewRGBA(canvas.Bounds())
			draw.Copy(backup, image.Point{}, canvas, canvas.Bounds(), draw.Src, nil)
		}

		// 현재 프레임을 캔버스에 합성 (투명 픽셀은 기존 화면 유지)
		draw.Draw(canvas, src.Bounds(), src, src.Bounds().Min, draw.Over)

		// 스냅샷 저장 (이후 캔버스가 변해도 프레임은 보존)
		snap := image.NewRGBA(canvas.Bounds())
		draw.Copy(snap, image.Point{}, canvas, canvas.Bounds(), draw.Src, nil)
		anim.Frames = append(anim.Frames, snap)

		// delay: GIF는 1/100초 단위, 0이면 기본 100ms
		d := time.Duration(g.Delay[i]) * 10 * time.Millisecond
		if d <= 0 {
			d = 100 * time.Millisecond
		}
		anim.Delays = append(anim.Delays, d)

		// 다음 프레임을 위한 disposal 적용
		switch g.Disposal[i] {
		case gif.DisposalBackground:
			// 프레임 영역을 투명(배경)으로 지움
			draw.Draw(canvas, src.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			if backup != nil {
				draw.Copy(canvas, image.Point{}, backup, backup.Bounds(), draw.Src, nil)
			}
		}
	}

	return anim, nil
}

// Resize는 이미지를 비율을 유지하며 targetW x targetH 로 리사이즈한다.
// targetH는 반블록 렌더링 기준으로 픽셀 높이 (터미널 행 * 2).
func Resize(src image.Image, targetW, targetH int) image.Image {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 비율 유지
	scaleW := float64(targetW) / float64(srcW)
	scaleH := float64(targetH) / float64(srcH)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}

	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	// 높이는 짝수 필요 (반블록: 2픽셀 = 1행)
	if newH%2 != 0 {
		newH++
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)
	return dst
}
