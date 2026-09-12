# tdraw 기여 가이드 (Contributing Guide)

`tdraw` 프로젝트에 관심을 가져주셔서 감사합니다! 버그 리포트, 기능 제안, 문서 개선, 코드 기여 등 모든 형태의 기여를 환영합니다.

---

## 1. 이슈 제보 및 기능 제안

- 버그를 발견하셨거나 새로운 기능을 제안하고 싶으신 경우 [GitHub Issues](https://github.com/wkqco33/tdraw/issues)에 등록해 주세요.
- 버그 제보 시 다음 정보를 포함해 주시면 빠른 해결에 도움이 됩니다:
  - OS 및 터미널 환경 (예: macOS Terminal, iTerm2, Windows Terminal 등)
  - `tdraw --version` 및 Go 버전 (`go version`)
  - 재현 방법 및 사용한 이미지 포맷/샘플
  - 기대한 동작과 실제 발생한 오류 메시지

---

## 2. 개발 환경 설정

`tdraw`는 크로스 플랫폼 빌드 도구인 [Task](https://taskfile.dev)를 사용합니다.

```bash
# 1. 저장소 복제
git clone https://github.com/wkqco33/tdraw.git
cd tdraw

# 2. 의존성 다운로드
go mod download

# 3. 바이너리 빌드
task build

# 4. 테스트 실행
task test
```

### 주요 Task 명령어

| 명령 | 설명 |
| :--- | :--- |
| `task build` | 기본 바이너리 빌드 (비디오 제외) |
| `task build:video` | 비디오 재생(`tcamviewer`) 포함 빌드 (CGO/FFmpeg 필요) |
| `task test` | 전체 단위 테스트 실행 |
| `task fmt` | `gofmt` 코드 포맷팅 |
| `task vet` | `go vet` 정적 분석 |
| `task clean` | 빌드 생성물 삭제 |

---

## 3. 개발 원칙 및 TDD

- **TDD(Test-Driven Development)**: 버그 수정이나 새 기능 구현 시 반드시 대응하는 단위 테스트(`*_test.go`)를 먼저 작성하거나 함께 커밋해야 합니다.
- **AI 에이전트 지침**: AI 에이전트와의 협업 가이드라인은 [`AGENTS.md`](AGENTS.md)를 참고하세요.
- **성능 유지**: 렌더링 핫루프를 수정하는 경우 `render/render_bench_test.go` 벤치마크를 실행해 메모리 할당(`allocs/op`) 증가가 없는지 확인하세요:
  ```bash
  go test -bench=. -benchmem ./render
  ```

---

## 4. 커밋 및 PR 컨벤션

`tdraw`는 [Conventional Commits](https://www.conventionalcommits.org/ko/v1.0.0/) 규격을 따릅니다:

- `feat`: 새로운 기능 추가
- `fix`: 버그 수정
- `docs`: 문서 변경 (README, CHANGELOG 등)
- `style`: 코드 포맷팅, 세미콜론 누락 등 (로직 변경 없음)
- `refactor`: 리팩터링 (기능 변경 없는 코드 구조 개선)
- `perf`: 성능 개선
- `test`: 테스트 코드 추가 및 수정
- `ci`: CI/CD 워크플로 변경

### PR 제출 전 체크리스트

1. [ ] `task fmt`를 실행해 코드 스타일을 맞추었는가?
2. [ ] `task vet` 검사를 통과했는가?
3. [ ] `task test` 실행 시 모든 테스트가 성공하는가?
4. [ ] 의존성 변경이 있는 경우 `govulncheck ./...` 검사를 통과했는가?
5. [ ] 사용자에게 영향을 주는 변경 사항을 [`CHANGELOG.md`](CHANGELOG.md)에 기록했는가?
