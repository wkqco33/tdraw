// Package tcam은 tcamviewer 라이브러리(FFmpeg 기반 C++ 코어)를
// video.FrameSource로 연결한다.
//
// 구현은 빌드 태그로 이원화되어 있다.
//
//   - cgo.go (`tcamviewer` 태그 + CGO): tcamviewer Go 패키지로 실제 디코딩.
//     libtcamviewer.so 가 필요하다(task build:video 참고).
//   - stub.go (그 외): ErrUnsupported 만 반환하는 스텁. 기본 빌드는 CGO 없이
//     순수 Go로 유지된다.
//
// 에디터의 gopls가 cgo.go를 보려면 빌드 플래그에 -tags=tcamviewer 가 필요하다.
// 본 저장소는 .pi-lens.json 에 해당 설정이 되어 있다.
package tcam

import "errors"

// ErrUnsupported는 비디오 재생 미지원 빌드(스텁)에서 Open이 반환하는 에러다.
var ErrUnsupported = errors.New(
	"이 빌드는 비디오 재생을 지원하지 않습니다. " +
		"libtcamviewer 를 빌드한 뒤 'task build:video' 로 다시 빌드하세요")
