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

## 요구사항

- Go 1.21+
- 터미널의 truecolor 지원 권장 (`COLORTERM=truecolor`)
