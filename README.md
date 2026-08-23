# tdraw

원격 접속(SSH 등) 환경처럼 터미널만 사용 가능한 상황에서 이미지를 확인하기 위한 CLI 뷰어.

Unicode half-block(`▀`)과 ANSI 24-bit 컬러를 이용해 터미널에 이미지를 출력한다.
GIF는 애니메이션으로 재생된다.

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
| `task build` | 현재 플랫폼용 바이너리 빌드 |
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
```

### 옵션

CLI는 [wcli](https://github.com/wkqco33/wcli) 프레임워크로 구현되어 있다. 긴 옵션은 `--`, 단축 옵션은 `-`를 사용한다.

| 옵션 | 기본값 | 설명 |
| ---- | ------ | ---- |
| `-w`, `--width int` | 터미널 너비 | 출력 너비 (열 수) |
| `-c`, `--color string` | `truecolor` | 컬러 모드: `truecolor` \| `256` \| `gray` |
| `--version` | - | 버전 출력 |
| `-h`, `--help` | - | 도움말 출력 |

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
