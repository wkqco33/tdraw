package imgindex

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

	_ "github.com/wkqco33/tdraw/pnm" // PBM/PGM/PPM (P1-P6)
)

type decodedInfo struct {
	format        string
	width, height int
}

func load(path string) (*decodedInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()
	var img image.Image
	var format string
	if strings.EqualFold(filepath.Ext(path), ".gif") {
		gifInfo, err := gif.DecodeConfig(f)
		if err != nil {
			return nil, fmt.Errorf("GIF 디코딩 실패: %w", err)
		}
		return &decodedInfo{format: "gif", width: gifInfo.Width, height: gifInfo.Height}, nil
	}
	img, format, err = image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("이미지 디코딩 실패: %w", err)
	}
	bounds := img.Bounds()
	return &decodedInfo{format: format, width: bounds.Dx(), height: bounds.Dy()}, nil
}
