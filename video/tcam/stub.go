//go:build !tcamviewer || !cgo

// tcam 패키지의 스텁 구현이다. `tcamviewer` 빌드 태그가 없거나 CGO가
// 비활성화된 빌드에서 컴파일되며, Open은 항상 안내 에러를 반환한다.
// 기본 빌드는 CGO 없이 순수 Go로 유지된다.
package tcam

import (
	"io"

	"github.com/wkqco33/tdraw/video"
)

// Source는 스텁 소스다. Open이 항상 실패하므로 생성되지 않는다.
type Source struct{}

// Open은 항상 ErrUnsupported를 반환한다.
func Open(path string, loop bool) (video.FrameSource, error) {
	return nil, ErrUnsupported
}

// 아래 메서드들은 Source가 video.FrameSource를 만족하도록 하는 자리 채움이다.
func (s *Source) Spec() video.Spec { return video.Spec{} }

func (s *Source) Rotation() int { return 0 }

func (s *Source) Next(targetW, targetH int) (rgb []byte, w, h, stride int, err error) {
	return nil, 0, 0, 0, io.EOF
}

func (s *Source) Rewind() error { return nil }

func (s *Source) Close() {}
