# tdraw

원격 접속(SSH 등) 환경처럼 터미널만 사용 가능한 상황에서 이미지를 확인하기 위한 CLI 뷰어.

Unicode half-block(`▀`)과 ANSI 24-bit 컬러를 이용해 터미널에 이미지를 출력한다.

## 렌더링 방식

- `▀` 문자 1개 = 픽셀 2개 (상단: foreground 색, 하단: background 색)
- ANSI truecolor (`ESC[38;2;R;G;Bm`) 사용으로 색상 그대로 표현
- 터미널 크기를 자동 감지해 비율 유지 리사이즈

## 설치

```bash
git clone <repo>
cd tdraw
go build -o tdraw .
```

또는 직접 실행:
```bash
go run . photo.jpg
```

## 사용법

```
tdraw [옵션] <이미지파일> [이미지파일...]
```

### 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-w int` | 터미널 너비 | 출력 너비 (열 수) |
| `-color string` | `truecolor` | 컬러 모드: `truecolor` \| `256` \| `gray` |
| `-version` | - | 버전 출력 |

### 예시

```bash
# 기본 출력 (터미널 크기 자동)
tdraw photo.jpg

# 너비 60열로 출력
tdraw -w 60 photo.jpg

# 흑백 모드
tdraw -color gray image.png

# 여러 파일 한번에
tdraw *.jpg

# 256색 모드 (truecolor 미지원 터미널)
tdraw -color 256 photo.jpg
```

## 지원 포맷

- JPEG, PNG, GIF (첫 프레임), WebP, BMP

## 컬러 모드

| 모드 | 설명 | 적합한 환경 |
|------|------|-------------|
| `truecolor` | 24-bit RGB, 1677만 색 | 최신 터미널 (iTerm2, Windows Terminal 등) |
| `256` | ANSI 256색 팔레트 | 구형 터미널, tmux 기본 설정 |
| `gray` | 회색조 | 색상 미지원 환경 |

## 알고리즘 및 기술적 특징

### Half-block 렌더링

터미널 문자 1개로 세로 2픽셀을 표현하는 핵심 기법이다.

- `▀` (U+2580) 문자의 **foreground** 색 → 상단 픽셀
- `▀` (U+2580) 문자의 **background** 색 → 하단 픽셀
- 이미지를 세로 2줄씩 읽어 한 문자에 압축하므로, 출력 높이 = 이미지 높이 ÷ 2

```
픽셀 행 0  →  ▀ (foreground)
픽셀 행 1  →  ▀ (background)
픽셀 행 2  →  ▀ (foreground)
픽셀 행 3  →  ▀ (background)
...
```

### 비율 유지 리사이즈

출력 너비와 터미널 높이를 목표 크기로 삼아 **가로·세로 독립 배율**을 계산한 뒤 작은 쪽을 선택한다.

```
scaleW = targetW / srcW
scaleH = targetH / srcH
scale  = min(scaleW, scaleH)   ← 이미지가 넘치지 않도록 작은 배율 사용
```

- half-block 렌더링 요구사항에 맞게 출력 높이를 항상 **짝수**로 강제
- 축소·확대 모두 **Bilinear Interpolation**으로 처리해 계단 현상 최소화

### 컬러 모드 구현

#### Truecolor (24-bit RGB)
ANSI 이스케이프 시퀀스로 1,677만 색을 그대로 표현한다.

```
ESC[38;2;R;G;Bm   ← foreground (상단 픽셀)
ESC[48;2;R;G;Bm   ← background (하단 픽셀)
```

#### 256색 모드
ANSI 256색 팔레트의 6×6×6 컬러 큐브(인덱스 16–231)에 매핑한다.

```
index = 16 + 36×r_i + 6×g_i + b_i
r_i   = R × 5 / 255   (0–5 범위로 양자화)
```

#### 회색조 모드
**ITU-R BT.601** 휘도 공식으로 인간의 색 인식 가중치를 반영한다.

```
Y = (299×R + 587×G + 114×B) / 1000
```

계산된 휘도를 ANSI 256색 팔레트의 회색 영역(인덱스 232–255)에 매핑한다.

### 터미널 크기 자동 감지

`golang.org/x/term` 패키지로 현재 터미널의 열·행 수를 런타임에 조회한다. 감지에 실패하면 80×24 기본값으로 폴백한다.

### 메모리 효율적 렌더링

`strings.Builder`로 한 줄 분량의 ANSI 시퀀스를 버퍼에 모은 뒤 한 번에 출력한다. 픽셀별로 개별 할당하지 않아 대용량 이미지도 안정적으로 처리된다.

## 요구사항

- Go 1.21+
- 터미널의 truecolor 지원 권장 (`COLORTERM=truecolor`)
