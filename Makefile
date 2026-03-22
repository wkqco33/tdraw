BIN     := tdraw
PREFIX  := $(HOME)/.local
BINDIR  := $(PREFIX)/bin
DISTDIR := dist

GO      := go
GOFLAGS := -trimpath
LDFLAGS := -s -w

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS += -X main.version=$(VERSION)

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: all build test clean install uninstall fmt vet release help

all: build

## build: 현재 플랫폼용 바이너리 빌드
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/tdraw

## test: 테스트 실행
test:
	$(GO) test ./... -v

## fmt: 코드 포맷
fmt:
	$(GO) fmt ./...

## vet: 정적 분석
vet:
	$(GO) vet ./...

## release: 모든 플랫폼용 바이너리를 dist/ 에 빌드
release: clean-dist
	@mkdir -p $(DISTDIR)
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(platform)))) \
		$(eval OUT  := $(DISTDIR)/$(BIN)_$(OS)_$(ARCH)$(if $(filter windows,$(OS)),.exe,)) \
		echo "빌드: $(OUT)" && \
		GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUT) ./cmd/tdraw && \
	) true
	@echo "완료: $(DISTDIR)/"
	@ls -lh $(DISTDIR)/

## install: $(BINDIR) 에 바이너리 설치
install: build
	@mkdir -p $(BINDIR)
	install -m 755 $(BIN) $(BINDIR)/$(BIN)
	@echo "설치 완료: $(BINDIR)/$(BIN)"

## uninstall: 설치된 바이너리 삭제
uninstall:
	rm -f $(BINDIR)/$(BIN)
	@echo "삭제 완료: $(BINDIR)/$(BIN)"

## clean: 빌드 결과물 삭제
clean:
	rm -f $(BIN)

## clean-dist: dist/ 디렉토리 삭제
clean-dist:
	rm -rf $(DISTDIR)

## help: 사용 가능한 타겟 목록 출력
help:
	@grep -E '^## ' Makefile | sed 's/^## /  /'
