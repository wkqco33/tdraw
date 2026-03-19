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

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
	"golang.org/x/image/draw"
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
