# Changelog

이 프로젝트의 모든 주요 변경 사항은 이 문서에 기록됩니다.

형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따르며,
버전 번호는 [Semantic Versioning](https://semver.org/lang/ko/)을 준수합니다.

## [Unreleased]

## [0.5.0] - 2026-09-13

### Added
- 전역 환경 설정 관리 패키지 (`config`): OS 표준 디렉터리 기반 JSON 설정 로드, 저장 및 점 표기법(`dot-notation`) 키/값 조회·수정.
- CLI `config` 서브커맨드 (`init`, `path`, `show`, `get`, `set`) 추가.
- 환경변수 `TDRAW_CONFIG`를 통한 설정 파일 경로 재정의 지원.
- 기본 플래그(`--width`, `--color`, `--quiet`, `--no-meta`) 및 서브커맨드(`play`, `find`, `ask`, `agent`, `ocr`, `index`)에 설정값 연동.
- `config` 패키지 및 서브커맨드 단위 테스트 추가.
- `wcli` CLI 프레임워크 라이브러리를 `v0.2.3`으로 업데이트.

## [0.4.0] - 2026-09-13

### Security
- `golang.org/x/image` 의존성을 `v0.46.0`으로 업데이트하여 보안 취약점 4건 해결 (GO-2026-6222, GO-2026-5061, GO-2026-5031, GO-2026-4961).

### Added
- 표준 입력(`-`)을 통한 파이프 이미지 수신 지원 (`cat photo.jpg | tdraw -`).
- `NO_COLOR` (https://no-color.org) 및 `TERM=dumb` 환경변수 감지 지원.
- 표준 CLI 플래그 추가: `--no-color`, `-q`/`--quiet`, `--no-meta`.
- `AGENTS.md`, `CONTRIBUTING.md`, `SECURITY.md`: 오픈소스 거버넌스 및 에이전트 협업 가이드라인 신설.
- `tdraw_test.go`, `cmd/tdraw/main_test.go`: 최상위 퍼사드 및 CLI 단위 테스트 추가 (퍼사드 커버리지 93.9% 달성).

### Changed
- `imgindex`: 대용량 이미지 인덱싱 시 전체 픽셀 디코딩 대신 `image.DecodeConfig`를 사용하여 메모리 사용량 및 속도 대폭 개선.
- `vision`: `imageMetadataTool`에서 불필요한 전체 이미지 디코딩 대신 헤더 메타데이터 전용 리더 사용.
- `cmd/tdraw`: 비TTY 환경에서 메타 박스 출력을 자동 생략하여 파이프/리다이렉션 스트림 격리.
- `cmd/tdraw`: 종료 코드 세분화 (정상 0, 런타임 오류 1, 인자/사용법 오류 2).
- `cmd/tdraw`: LLM API 요청에 60초 기본 타임아웃 컨텍스트 적용.
- `ci`: GitHub Actions 워크플로에 `gofmt` 코드 포맷 검사 스텝 추가.

## [0.3.0] - 2026-09-09

### Added
- 비디오 파일(MP4, MKV, AVI, WebM 등) 실시간 터미널 재생 엔진 (`video` 패키지).
- `tdraw play` 서브커맨드 추가 (`--loop`, `-w`, `--color`).
- `tcamviewer` (FFmpeg 기반 C++ 코어) CGO 바인딩 연동.
- Linux 릴리스 바이너리에 비디오 재생 기능 정적 링크 포함.

## [0.2.2] - 2026-09-08

### Added
- Vision API 전송 시 PNM(PBM/PGM/PPM) 이미지를 PNG로 자동 변환하여 전송하는 호환성 계층 추가.

## [0.2.1] - 2026-09-08

### Added
- Netpbm PNM 계열 포맷(PBM P1/P4, PGM P2/P5, PPM P3/P6) 디코더 (`pnm` 패키지).
- `image.RegisterFormat`을 통한 표준 `image.Decode` 연동.

## [0.2.0] - 2026-09-07

### Added
- Ollama 기반 Vision 모델 이미지 질의 커맨드 (`tdraw ask`).
- 읽기 전용 이미지 메타데이터 분석 에이전트 커맨드 (`tdraw agent`).
- 이미지 텍스트 추출 커맨드 (`tdraw ocr`).
- 로컬 이미지 Vision 설명 인덱싱 및 검색 (`tdraw index`, `tdraw find`).

## [0.1.5] - 2026-08-23

### Added
- PPM 패키지 매니저용 릴리스 자산 및 SHA-256 체크섬 생성 자동화 (`task release`).

## [0.1.4] - 2026-06-27

### Changed
- 렌더링 성능 최적화: `*image.RGBA` 대상 직렬화 핫루프 최적화 (`RenderString` 1 allocs/op 달성).

## [0.1.3] - 2026-06-26

### Added
- GitHub Actions 크로스 플랫폼(Linux, macOS, Windows) CI 빌드 및 테스트 자동화.

## [0.1.2] - 2026-06-21

### Added
- 종횡비 유지 Bilinear 리사이즈 및 Half-block 렌더링 개선.
- ANSI 256색 및 Grayscale(ITU-R BT.601) 컬러 변환 모드 지원.
- GIF 애니메이션 재생 및 disposal method 합성 지원.

## [0.1.1] - 2026-03-22

### Added
- `wcli` 기반의 CLI 기본 골격 구현.

## [0.1.0] - 2026-03-20

### Added
- 프로젝트 초기화.
