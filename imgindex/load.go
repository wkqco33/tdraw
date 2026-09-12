package imgindex

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

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

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, fmt.Errorf("이미지 디코딩 실패: %w", err)
	}
	return &decodedInfo{format: format, width: cfg.Width, height: cfg.Height}, nil
}
