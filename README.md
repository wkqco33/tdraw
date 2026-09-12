# tdraw

원격 접속(SSH 등) 환경처럼 터미널만 사용 가능한 상황에서 이미지를 확인하기 위한 CLI 뷰어.

Unicode half-block(`▀`)과 ANSI 24-bit 컬러를 이용해 터미널에 이미지를 출력한다.
GIF는 애니메이션으로 재생되고, 비디오 파일(MP4/MKV 등)은 `play` 명령으로 터미널에서
재생할 수 있다.

CLI로 직접 사용하거나, Go 라이브러리로 임포트해 사용할 수 있다.

## CLI 설치

```bash
go install github.com/wkqco33/tdraw/cmd/tdraw@latest
```

또는 소스에서 빌드. 빌드는 [Task](https://taskfile.dev)를 사용한다(크로스 플랫폼, Windows 포함).

```bash
# 소스에서 클론
git clone https://github.com/wkqco33/tdraw
cd tdraw
task build
```

주요 태스크 목록은 `task --list`로 확인한다.

| 태스크 | 설명 |
| ------ | ---- |
| `task build` | 현재 플랫폼용 바이너리 빌드 (비디오 재생 미포함) |
| `task build:video` | 비디오 재생(tcamviewer) 포함 빌드 |
| `task test` | 테스트 실행 |
| `task release` | PPM용 플랫폼별 아카이브와 SHA-256 체크섬을 `dist/` 에 생성 |
| `task install` | `~/.local/bin` 에 설치 (Unix 전용) |
| `task clean` | 빌드 결과물 삭제 |

`task release`는 PPM 패키지 매니저에서 자동 인식할 수 있는 다음 릴리스 자산을 생성한다.

```text
tdraw_linux_amd64.tar.gz
tdraw_linux_amd64.tar.gz.sha256
tdraw_linux_arm64.tar.gz
tdraw_linux_arm64.tar.gz.sha256
tdraw_darwin_amd64.tar.gz
tdraw_darwin_amd64.tar.gz.sha256
tdraw_darwin_arm64.tar.gz
tdraw_darwin_arm64.tar.gz.sha256
tdraw_windows_amd64.zip
tdraw_windows_amd64.zip.sha256
```

GitHub Release에 `dist/`의 아카이브와 체크섬 파일을 함께 업로드하면 PPM으로 설치할 수 있다.

## CLI 사용법

```bash
tdraw [옵션] <이미지파일> [이미지파일...]

# 표준 입력(stdin)에서 이미지 읽기
cat photo.jpg | tdraw -
curl -sL https://example.com/sample.png | tdraw -
```

### 옵션

CLI는 [wcli](https://github.com/wkqco33/wcli) 프레임워크로 구현되어 있다. 긴 옵션은 `--`, 단축 옵션은 `-`를 사용한다.

| 옵션 | 기본값 | 설명 |
| ---- | ------ | ---- |
| `-w`, `--width int` | 터미널 너비 | 출력 너비 (열 수) |
| `-c`, `--color string` | `truecolor` | 컬러 모드: `truecolor` \| `256` \| `gray` |
| `--no-color` | - | 컬러 출력 비활성화 (`gray` 모드 적용, `NO_COLOR` 환경변수 지원) |
| `-q`, `--quiet` | - | 진행률 및 메타데이터 출력 억제 |
| `--no-meta` | - | 상단 메타데이터 정보 박스 출력 생략 (비TTY 리다이렉션 시 자동 생략) |
| `-V`, `--verbose` | - | 상세 진단 로그 출력 (stderr) |
| `--version` | - | 버전 출력 |
| `-h`, `--help` | - | 도움말 출력 |

### AI 이미지 질의

Ollama의 Vision 모델을 사용해 이미지에 자연어 질문을 할 수 있다. 기본 모델은
`llava`이며, Ollama가 실행 중이고 모델이 설치되어 있어야 한다.

```bash
# Ollama 설치 후 Vision 모델 준비
ollama pull llava

# 이미지 설명
tdraw ask photo.jpg "이 이미지에 무엇이 보이나요?"

# 스크린샷의 오류 메시지 확인
tdraw ask screenshot.png "오류 메시지만 간결하게 알려줘"

# JSON 출력
tdraw ask photo.jpg "주요 객체를 알려줘" --json

# 이미지 텍스트 추출
tdraw ocr screenshot.png
tdraw ocr receipt.jpg --json
```

`ask`는 다음 환경변수로 Ollama 연결을 설정할 수 있다.

| 환경변수 | 기본값 | 설명 |
| -------- | ------ | ---- |
| `TDRAW_LLM_MODEL` | `llava` | 사용할 Ollama Vision 모델 |
| `TDRAW_OLLAMA_URL` | `http://localhost:11434/v1` | Ollama OpenAI 호환 API 주소 |

플래그로도 설정할 수 있다.

```bash
tdraw ask --model llama3.2-vision image.jpg "이미지를 설명해줘"
tdraw ask --ollama-url http://localhost:11434/v1 image.jpg "텍스트를 읽어줘"
```

이미지는 분석을 위해 Ollama 서버로 전송된다. 민감한 이미지에는 로컬에서 실행되는
Ollama를 사용하고, Ollama Vision 모델이 아닌 텍스트 전용 모델은 사용할 수 없다.
PBM/PGM/PPM 이미지는 Vision API 호환성을 위해 요청 전에 PNG로 변환되어 전송된다.

### AI 에이전트

`agent` 명령은 이미지와 요청을 Ollama Vision 모델에 전달하고, 필요한 경우 이미지의
정확한 포맷과 크기를 읽기 전용 도구로 확인한다. 현재 파일을 수정하거나 외부 명령을
실행하는 도구는 제공하지 않는다.

```bash
tdraw agent screenshot.png "이미지 내용을 요약하고 원본 크기도 알려줘"
tdraw agent photo.jpg "사람이 있는지와 이미지 크기를 알려줘" --json
```

`agent`도 `TDRAW_LLM_MODEL`, `TDRAW_OLLAMA_URL`, `--model`, `--ollama-url` 설정을
`ask`와 동일하게 사용한다.

### OCR

`ocr`는 이미지의 텍스트만 추출하며, 원래 줄바꿈과 읽기 순서를 유지하도록
Vision 모델에 요청한다.

```bash
tdraw ocr terminal-error.png
tdraw ocr document.jpg --model llama3.2-vision
tdraw ocr receipt.png --json
```

OCR 결과는 모델 응답에 의존하므로 작은 글씨, 흐린 이미지, 손글씨에서는 정확도가
낮을 수 있다. 출력이 필요하면 `--json`으로 이미지 경로와 모델명을 함께 받을 수 있다.

### 이미지 인덱싱과 검색

여러 이미지를 반복해서 검색할 때는 먼저 Vision 설명을 로컬 인덱스에 저장한다.
이미지 원본은 인덱스에 복사하지 않는다.

```bash
# 기본 출력: ./photos/.tdraw/index.json
tdraw index ./photos

# 설명과 경로로 검색
tdraw find ./photos "터미널 오류"
tdraw find ./photos "고양이" --limit 10

# 자동화용 JSON 출력
tdraw find ./photos "영수증" --json
```

인덱스 생성 시 각 이미지가 Ollama로 전송되며, 생성된 설명과 포맷·크기만 JSON으로
저장된다. 검색 단계에서는 모델을 호출하지 않고 로컬 인덱스만 사용한다. 출력 위치는
`tdraw index ./photos --output /path/to/index.json`, 검색 인덱스는
`tdraw find ./photos "검색어" --index /path/to/index.json`으로 변경할 수 있다.

### 예시

```bash
# 기본 출력 (터미널 크기 자동)
tdraw photo.jpg

# 너비 60열로 출력
tdraw -w 60 photo.jpg

# 흑백 모드
tdraw --color gray image.png

# 여러 파일 한번에
tdraw *.jpg

# 256색 모드 (truecolor 미지원 터미널)
tdraw --color 256 photo.jpg

# GIF 애니메이션 재생 (Ctrl+C로 종료)
tdraw anim.gif
```

### 비디오 재생

MP4, MKV, AVI, WebM 등 [FFmpeg](https://ffmpeg.org)이 지원하는 비디오 파일을
터미널에서 실시간 재생한다. 디코딩은 [tcamviewer](https://github.com/wkqco33/tcamviewer)
라이브러리(FFmpeg 기반 C++ 코어)가 담당한다. 오디오는 출력되지 않는다.

```bash
tdraw play video.mp4

# 스트림 끝에서 처음부터 반복 재생
tdraw play --loop video.mp4

# 너비 60열, 256색 모드
tdraw play -w 60 --color 256 video.mp4
```

비디오 재생은 CGO가 필요하다. **릴리스 Linux 바이너리에는 비디오 재생이 포함되어
있으며**, FFmpeg 런타임 라이브러리가 필요하다(대부분의 배포판에 기본 설치되어
있다. 없다면 `sudo apt install ffmpeg`).

소스에서 빌드하려면 먼저 tcamviewer 코어 라이브러리를 빌드한 뒤
`task build:video` 를 실행한다. `task build:video` 는 로컬 tcamviewer
체크아웃을 go.work로 자동 연결한다(go.work 는 커밋되지 않는다).

```bash
# 1. tcamviewer 코어 라이브러리 빌드 (../tcamviewer 위치)
git clone https://github.com/wkqco33/tcamviewer ../tcamviewer
cd ../tcamviewer && task build:core

# 2. 비디오 재생 포함 tdraw 빌드
cd ../tdraw && task build:video
```

`go install` 바이너리와 macOS/Windows 릴리스 바이너리는 비디오 재생이 빠져
있으며, `tdraw play` 실행 시 안내 메시지가 출력된다.

### 셸 자동 완성

```bash
# Bash
source <(tdraw completion bash)

# Zsh
source <(tdraw completion zsh)

# Fish
tdraw completion fish > ~/.config/fish/completions/tdraw.fish
```

## 라이브러리 사용법

```bash
go get github.com/wkqco33/tdraw
```

### API

```go
import "github.com/wkqco33/tdraw"

// 파일 경로로 렌더링
err := tdraw.DrawFile(os.Stdout, "photo.jpg", tdraw.Options{
    Width:     80,
    ColorMode: tdraw.TrueColor,
})

// image.Image로 렌더링
tdraw.Draw(os.Stdout, img, tdraw.Options{})

// GIF 애니메이션 재생 (ctx 취소 시 종료)
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
tdraw.PlayGIFFile(ctx, os.Stdout, "anim.gif", tdraw.Options{})

// 터미널 크기 조회
cols, rows := tdraw.TermSize()
```

### Options

| 필드 | 타입 | 기본값 | 설명 |
| ---- | ------ | -------- | ------ |
| `Width` | `int` | 터미널 너비 자동 감지 | 0이면 자동 |
| `Height` | `int` | `Width * 4` | 0이면 자동 (비율 유지) |
| `ColorMode` | `ColorMode` | `TrueColor` | `TrueColor` \| `Color256` \| `Gray` |

### 구현 예시

```go
package main

import (
    "os"
    "github.com/wkqco33/tdraw"
)

func main() {
    // 터미널 너비 자동 감지, TrueColor
    tdraw.DrawFile(os.Stdout, "photo.jpg", tdraw.Options{})

    // 너비 고정, 256색
    tdraw.DrawFile(os.Stdout, "photo.jpg", tdraw.Options{
        Width:     60,
        ColorMode: tdraw.Color256,
    })
}
```

## 지원 포맷

- JPEG, PNG, GIF (애니메이션 재생), WebP, BMP
- PNM 계열: PGM (P2/P5, 회색조), PPM (P3/P6, 컬러), PBM (P1/P4, 흑백)

## 컬러 모드

| 모드 | 설명 | 적합한 환경 |
| ---- | ------ | ------------- |
| `truecolor` / `TrueColor` | 24-bit RGB, 1677만 색 | 최신 터미널 (iTerm2, Windows Terminal 등) |
| `256` / `Color256` | ANSI 256색 팔레트 | 구형 터미널, tmux 기본 설정 |
| `gray` / `Gray` | 회색조 | 색상 미지원 환경 |

## 알고리즘 및 기술적 특징

### Half-block 렌더링

터미널 문자 1개로 세로 2픽셀을 표현하는 핵심 기법이다.

- `▀` (U+2580) 문자의 **foreground** 색 → 상단 픽셀
- `▀` (U+2580) 문자의 **background** 색 → 하단 픽셀
- 이미지를 세로 2줄씩 읽어 한 문자에 압축하므로, 출력 높이 = 이미지 높이 ÷ 2

```bash
픽셀 행 0  →  ▀ (foreground)
픽셀 행 1  →  ▀ (background)
픽셀 행 2  →  ▀ (foreground)
픽셀 행 3  →  ▀ (background)
...
```

### 비율 유지 리사이즈

출력 너비와 터미널 높이를 목표 크기로 삼아 **가로·세로 독립 배율**을 계산한 뒤 작은 쪽을 선택한다.

```bash
scaleW = targetW / srcW
scaleH = targetH / srcH
scale  = min(scaleW, scaleH)   ← 이미지가 넘치지 않도록 작은 배율 사용
```

- half-block 렌더링 요구사항에 맞게 출력 높이를 항상 **짝수**로 강제
- 축소·확대 모두 **Bilinear Interpolation**으로 처리해 계단 현상 최소화

### 컬러 모드 구현

#### Truecolor (24-bit RGB)

ANSI 이스케이프 시퀀스로 1,677만 색을 그대로 표현한다.

```bash
ESC[38;2;R;G;Bm   ← foreground (상단 픽셀)
ESC[48;2;R;G;Bm   ← background (하단 픽셀)
```

#### 256색 모드

ANSI 256색 팔레트의 6×6×6 컬러 큐브(인덱스 16–231)에 매핑한다.

```bash
index = 16 + 36×r_i + 6×g_i + b_i
r_i   = R × 5 / 255   (0–5 범위로 양자화)
```

#### 회색조 모드

**ITU-R BT.601** 휘도 공식으로 인간의 색 인식 가중치를 반영한다.

```bash
Y = (299×R + 587×G + 114×B) / 1000
```

계산된 휘도를 ANSI 256색 팔레트의 회색 영역(인덱스 232–255)에 매핑한다.

### 터미널 크기 자동 감지

`golang.org/x/term` 패키지로 현재 터미널의 열·행 수를 런타임에 조회한다. 감지에 실패하면 80×24 기본값으로 폴백한다.

### 메모리 효율적 렌더링

`strings.Builder`로 한 줄 분량의 ANSI 시퀀스를 버퍼에 모은 뒤 한 번에 출력한다. 픽셀별로 개별 할당하지 않아 대용량 이미지도 안정적으로 처리된다.

### GIF 애니메이션 재생

GIF 프레임은 전체 캔버스의 일부 영역만 담고 있어, disposal method에 따라 누적 합성해야 실제 화면이 된다.

- **합성**: 프레임을 캔버스에 `draw.Over`로 그린 뒤 스냅샷을 저장. disposal이 `Background`면 해당 영역을 투명으로 지우고, `Previous`면 직전 캔버스 상태로 복원
- **재생**: 합성된 프레임을 미리 문자열로 렌더링해 캐싱한 뒤, 커서를 위로 이동(`ESC[nA`)시키며 같은 위치에 덮어써 깜빡임 없이 재생
- **delay**: GIF의 1/100초 단위 지연을 프레임별로 반영 (0이면 100ms)
- **종료**: 재생 중 커서를 숨기고(`ESC[?25l`), `context` 취소(Ctrl+C) 시 커서를 복원(`ESC[?25h`)한 뒤 종료

## 요구사항

- Go 1.26.1+
- 터미널의 truecolor 지원 권장 (`COLORTERM=truecolor`)

## 버전 및 기여 가이드

- **버전 규칙**: [Semantic Versioning 2.0.0](https://semver.org/lang/ko/)을 준수합니다. 현재 `0.y.z` 단계로 v1.0.0 이전까지는 공개 API가 안정화되는 과정에서 일부 변경될 수 있습니다.
- **변경 이력**: 버전별 상세 변경 내역은 [CHANGELOG.md](CHANGELOG.md)를 참고하세요.
- **기여 가이드**: 버그 제보, 기능 제안 및 풀 리퀘스트 작성 지침은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참고하세요.
- **보안 정책**: 취약점 제보 방법은 [SECURITY.md](SECURITY.md)를 참고하세요.
- **AI 에이전트 지침**: 코드베이스 아키텍처 및 TDD 개발 규칙은 [AGENTS.md](AGENTS.md)를 참고하세요.

