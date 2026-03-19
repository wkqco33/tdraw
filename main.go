package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tdraw/imgutil"
	"tdraw/render"
)

// version은 빌드 시 ldflags로 주입된다: -X main.version=vX.Y.Z
var version = "dev"

func main() {
	var (
		width     = flag.Int("w", 0, "출력 너비 (열 수, 기본값: 터미널 너비)")
		colorMode = flag.String("color", "truecolor", "컬러 모드: truecolor | 256 | gray")
		showVer   = flag.Bool("version", false, "버전 출력")
	)
	flag.Usage = usage
	flag.Parse()

	if *showVer {
		fmt.Println("tdraw v" + version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	mode := parseColorMode(*colorMode)

	cols, rows := render.TermSize()
	_ = rows
	// 폭은 터미널 폭을 초과하면 줄바꿈이 발생하므로 cols로 제한
	targetW := cols
	if *width > 0 {
		targetW = *width
	}
	// targetH를 크게 잡아 이미지가 비율에 맞게 최대한 확대되도록 함
	// (세로는 스크롤 가능하므로 제한 없이 품질 우선)
	targetH := cols * 4

	exitCode := 0
	for i, path := range args {
		if i > 0 {
			fmt.Println()
		}
		if err := showImage(path, targetW, targetH, mode); err != nil {
			fmt.Fprintf(os.Stderr, "오류 [%s]: %v\n", filepath.Base(path), err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func showImage(path string, targetW, targetH int, mode render.ColorMode) error {
	info, err := imgutil.Load(path)
	if err != nil {
		return err
	}

	resized := imgutil.Resize(info.Image, targetW, targetH)
	rb := resized.Bounds()

	fmt.Printf("📄 %s  [%s]  원본: %dx%d  출력: %dx%d\n",
		filepath.Base(path),
		strings.ToUpper(info.Format),
		info.Width, info.Height,
		rb.Dx(), rb.Dy()/2, // /2: 반블록이므로 실제 터미널 행 수
	)

	render.Render(os.Stdout, resized, mode)
	return nil
}

func parseColorMode(s string) render.ColorMode {
	switch strings.ToLower(s) {
	case "256":
		return render.Color256
	case "gray", "grey":
		return render.ColorGray
	default:
		return render.ColorTruecolor
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `사용법: tdraw [옵션] <이미지파일> [이미지파일...]

옵션:
  -w int        출력 너비 (열 수). 기본값: 터미널 너비 자동 감지
  -color string 컬러 모드: truecolor (기본) | 256 | gray
  -version      버전 출력

지원 포맷: JPEG, PNG, GIF, WebP, BMP

예시:
  tdraw photo.jpg
  tdraw -w 60 photo.jpg
  tdraw -color gray image.png
  tdraw *.jpg
`)
}
