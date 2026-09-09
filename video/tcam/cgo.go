//go:build tcamviewer && cgo

// Package tcam은 tcamviewer 라이브러리(FFmpeg 기반 C++ 코어)를
// video.FrameSource로 연결한다.
//
// 이 파일은 `tcamviewer` 빌드 태그가 있고 CGO가 활성화된 환경에서만
// 컴파일된다. libtcamviewer.so가 필요하다(README의 build:video 참고).
// 에디터의 gopls가 이 파일을 보려면 빌드 플래그에 -tags=tcamviewer 가
// 필요하다. 본 저장소는 .pi-lens.json 에 설정되어 있다.
package tcam

import (
	"io"

	tcv "github.com/wkqco33/tcamviewer/pkg/tcamviewer"

	"github.com/wkqco33/tdraw/video"
)

// Source는 tcamviewer 디코더로 프레임을 읽어오는 FrameSource 구현이다.
type Source struct {
	dec  *tcv.Decoder
	spec video.Spec
}

// Open은 비디오 파일(또는 RTSP/웹캠 등 소스 문자열)을 열어 FrameSource를
// 반환한다. loop는 스트림 끝에서 되감아 반복할지를 뜻한다.
func Open(path string, loop bool) (video.FrameSource, error) {
	dec, err := tcv.NewDecoder(path, loop)
	if err != nil {
		return nil, err
	}
	w, h, fps, err := dec.Info()
	if err != nil {
		dec.Close()
		return nil, err
	}
	return &Source{
		dec:  dec,
		spec: video.Spec{Width: w, Height: h, FPS: fps},
	}, nil
}

func (s *Source) Spec() video.Spec { return s.spec }

func (s *Source) Rotation() int { return s.dec.Rotation() }

func (s *Source) Next(targetW, targetH int) (rgb []byte, w, h, stride int, err error) {
	f, err := s.dec.ReadFrame(targetW, targetH)
	if err != nil {
		// io.EOF는 그대로 전달해 스트림 끝을 알린다.
		if err == io.EOF {
			return nil, 0, 0, 0, io.EOF
		}
		return nil, 0, 0, 0, err
	}
	return f.Data, f.Width, f.Height, f.Stride, nil
}

func (s *Source) Rewind() error { return s.dec.Rewind() }

func (s *Source) Close() { s.dec.Close() }
