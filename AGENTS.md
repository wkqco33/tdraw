# AGENTS.md

`tdraw` 프로젝트에 기여하는 AI 에이전트 및 개발자를 위한 아키텍처 및 개발 지침서입니다.

---

## 1. 프로젝트 개요 및 아키텍처

`tdraw`는 원격 접속(SSH 등) 환경에서 터미널을 통해 이미지를 확인하기 위한 Go CLI 뷰어이자 라이브러리입니다. Unicode half-block(`▀`)과 ANSI 24-bit/256색 시퀀스를 사용하여 터미널 셀당 세로 2픽셀을 렌더링합니다.

### 패키지 구조 및 책임 분리

- **`github.com/wkqco33/tdraw` (루트)**: 외부에 노출되는 최상위 퍼사드(Facade) API. `Draw`, `DrawFile`, `PlayGIFFile`, `TermSize` 제공.
- **`cmd/tdraw`**: CLI 엔트리포인트 (`wcli` 프레임워크 기반). 플래그 파싱, 표준 입력(`-`) 처리, 메타데이터 출력 제어, 서브커맨드(`ask`, `agent`, `ocr`, `index`, `find`, `play`).
- **`render`**: Half-block 기반 텍스트 버퍼 렌더링 엔진. 성능 크리티컬한 핫루프이므로 메모리 할당을 최소화(1 allocs/op)하도록 설계됨.
- **`imgutil`**: 이미지 로딩, GIF 누적 disposal 합성, 종횡비 유지 Bilinear 리사이즈.
- **`pnm`**: Netpbm 계열(PBM/PGM/PPM, P1~P6) 이미지 표준 디코더. `image.RegisterFormat`을 통해 `image.Decode`와 연동.
- **`video`**: 비디오 프레임 스트림 실시간 터미널 재생 엔진. `video/tcam` CGO 바인딩을 통해 `tcamviewer` (FFmpeg)와 연동.
- **`vision`**: Ollama Vision 모델 기반 이미지 질의(`AskFile`), 에이전트(`RunAgentFile`), `OCRFile`.
- **`imgindex`**: 로컬 이미지 디렉터리 대상 Vision 설명 JSON 인덱스 생성 및 키워드 검색.

---

## 2. 개발 및 테스트 가이드 (TDD 원칙)

에이전트는 모든 기능 추가 및 버그 수정 시 **TDD(Test-Driven Development)** 원칙을 따릅니다.

1. **테스트 우선 작성**: 새로운 기능이나 변경 사항을 구현하기 전 실패하는 테스트(`*_test.go`)를 먼저 작성하거나 업데이트합니다.
2. **단위 테스트 독립성**: 외부 서비스(Ollama 서버 등)에 의존하는 테스트는 모의(Mock) 객체(`vision.Completer` 등 인터페이스)를 사용해 네트워크 없이 즉시 통과할 수 있어야 합니다.
3. **핫루프 성능 보호**: 렌더링이나 디코딩 등 빈번하게 호출되는 코드를 변경할 때는 `render/render_bench_test.go`와 같은 벤치마크를 실행하여 할당 수(`allocs/op`)와 소요 시간 회귀가 없는지 확인합니다:
   ```bash
   go test -bench=. -benchmem ./render
   ```
4. **전체 테스트 실행**:
   ```bash
   task test
   # 또는
   go test ./... -v
   ```

---

## 3. 코드 스타일 및 품질 관리

- **포맷팅**: 모든 코드는 `gofmt` 표준을 엄격히 준수해야 합니다.
  ```bash
  task fmt  # go fmt ./...
  ```
- **정적 분석**: 커밋 전 항상 `go vet`을 통과해야 합니다.
  ```bash
  task vet  # go vet ./...
  ```
- **주석 및 문서화**:
  - Export된 함수, 타입, 상수는 반드시 Go 표준 형태의 Doc comment를 작성합니다 (예: `// Draw renders...`).
  - 불필요한 독백, 에이전트 사설, 장황한 설명 주석은 금지하며 코드의 의도와 비직관적인 맥락만 간결하게 기술합니다.
- **성능 고려**: 이미지의 크기나 포맷 정보만 필요한 경우 전체 픽셀을 메모리에 로드하는 `image.Decode` 대신 `image.DecodeConfig`를 사용합니다.

---

## 4. CLI 가이드라인 준수 원칙

- **입출력 분리**:
  - 사용자에게 의미 있는 결과(렌더링된 이미지 스트림, 검색 결과, 질의 답변)는 `stdout`으로 출력합니다.
  - 진행 상태(스피너, 프로그레스 바), 진단 로그, 에러 메시지는 반드시 `stderr`로 출력합니다.
  - 리다이렉션 또는 파이프 연결 시 메타 박스 등의 부가 정보가 이미지 스트림을 오염시키지 않도록 제어해야 합니다.
- **터미널 환경 적응**:
  - [`NO_COLOR`](https://no-color.org) 환경변수가 설정되어 있거나 `TERM=dumb`인 경우 ANSI 컬러 이스케이프 시퀀스와 애니메이션/스피너 출력을 억제합니다.
  - 터미널이 아닌 환경(비TTY 파이프)에서는 대화형 프롬프트나 스피너를 출력하지 않습니다.
- **종료 코드 (Exit Codes)**:
  - 정상 종료: `0`
  - 일반 실행/런타임 실패: `1`
  - 잘못된 인자 또는 사용법 오류: `2`

---

## 5. 의존성 및 보안 정책

- 외부 의존성은 최소한으로 유지하며 표준 라이브러리를 우선합니다.
- 라이브러리 추가 또는 버전 변경 시 취약점을 점검합니다:
  ```bash
  govulncheck ./...
  ```
- 비밀키, 토큰, 암호화 키 등 민감 정보가 Git 히스토리나 로그에 남지 않도록 주의합니다.
