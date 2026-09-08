package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/wkqco33/LLM_client_go/ollama"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/logging"
	"github.com/wkqco33/wcli/rich"
	"golang.org/x/term"

	"github.com/wkqco33/tdraw/imgindex"
	"github.com/wkqco33/tdraw/imgutil"
	"github.com/wkqco33/tdraw/render"
	"github.com/wkqco33/tdraw/vision"
)

// version은 빌드 시 ldflags로 주입된다: -X main.version=vX.Y.Z
var version = "dev"

// heightScaleFactor는 자동 높이 상한을 출력 폭의 배수로 정한다.
// 세로는 스크롤 가능하므로 비율 유지 확대 시 품질을 우선해 넉넉히 잡는다.
const heightScaleFactor = 4

// errFailed는 처리 실패를 알리는 센티넬 에러다. SilenceErrors로 자동 출력이
// 억제되므로 메시지는 호출부에서 직접 출력하고, main은 종료 코드 판정에만 사용한다.
var errFailed = errors.New("처리 실패")

func main() {
	var (
		width      int
		colorMode  string
		verbose    bool
		askModel   string
		ollamaURL  string
		jsonOut    bool
		agentModel string
		agentURL   string
		agentJSON  bool
		ocrModel   string
		ocrURL     string
		ocrJSON    bool
		indexModel string
		indexURL   string
		indexOut   string
		findIndex  string
		findLimit  int
		findJSON   bool
	)

	root := &wcli.Command{
		Use:     "tdraw [옵션] <이미지파일>...",
		Short:   "이미지를 터미널에 렌더링하는 CLI",
		Long:    "JPEG, PNG, GIF, WebP, BMP, PGM(PPM/PBM) 이미지를 터미널에 컬러로 출력한다.\nGIF는 애니메이션으로 무한 반복 재생하며 Ctrl+C로 종료한다.",
		Version: "tdraw " + version,
		// 파일별 에러는 직접 출력하므로 wcli의 자동 에러 출력을 끈다.
		SilenceErrors: true,
		Run: func(ctx *wcli.Context) error {
			setupLogger(verbose)
			return run(ctx, width, colorMode)
		},
	}

	root.Flags().IntVar(&width, "width", "w", 0, "출력 너비 (열 수, 기본값: 터미널 너비)")
	root.Flags().StringVar(&colorMode, "color", "c", "truecolor", "컬러 모드: truecolor | 256 | gray")
	root.Flags().BoolVar(&verbose, "verbose", "V", false, "상세 진단 로그 출력 (stderr)")
	root.Flags().SetValidation("color", validateColorMode)

	root.AddCommand(wcli.NewCompletionCommand(root))
	root.AddCommand(newAskCommand(&askModel, &ollamaURL, &jsonOut))
	root.AddCommand(newAgentCommand(&agentModel, &agentURL, &agentJSON))
	root.AddCommand(newOCRCommand(&ocrModel, &ocrURL, &ocrJSON))
	root.AddCommand(newIndexCommand(&indexModel, &indexURL, &indexOut))
	root.AddCommand(newFindCommand(&findIndex, &findLimit, &findJSON))

	if err := root.Execute(os.Args[1:]); err != nil {
		// 파일별 상세 에러는 run()에서 이미 출력했으므로(errFailed),
		// 그 외 프레임워크 레벨 에러(플래그 파싱/검증 등)만 여기서 출력한다.
		if !errors.Is(err, errFailed) {
			rich.Fprintln(os.Stderr, "[red][bold]오류:[/bold] %s[/red]", err.Error())
		}
		os.Exit(1)
	}
}

func newAskCommand(model, ollamaURL *string, jsonOut *bool) *wcli.Command {
	ask := &wcli.Command{
		Use:   "ask <이미지파일> <질문>",
		Short: "Ollama Vision 모델로 이미지에 질문",
		Long:  "Ollama의 Vision 모델에 이미지를 전달하고 자연어 질문을 수행한다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) < 2 {
				return errors.New("이미지 파일과 질문을 지정하세요 (예: tdraw ask photo.jpg \"무엇이 보이나요?\")")
			}

			configuredModel := *model
			if configuredModel == "" {
				configuredModel = envOrDefault("TDRAW_LLM_MODEL", "llava")
			}
			configuredURL := *ollamaURL
			if configuredURL == "" {
				configuredURL = envOrDefault("TDRAW_OLLAMA_URL", "http://localhost:11434/v1")
			}

			client := ollama.New(ollama.Config{BaseURL: configuredURL})
			question := strings.Join(ctx.Args[1:], " ")
			answer, err := vision.AskFile(ctx.Context, client, configuredModel, ctx.Args[0], question)
			if err != nil {
				return err
			}
			if *jsonOut {
				fmt.Printf("{\"image\":%q,\"model\":%q,\"answer\":%q}\n", ctx.Args[0], configuredModel, answer)
				return nil
			}
			fmt.Println(answer)
			return nil
		},
	}
	ask.Flags().StringVar(model, "model", "m", "", "Ollama Vision 모델 (기본값: TDRAW_LLM_MODEL 또는 llava)")
	ask.Flags().StringVar(ollamaURL, "ollama-url", "", "", "Ollama OpenAI 호환 주소 (기본값: TDRAW_OLLAMA_URL 또는 localhost:11434/v1)")
	ask.Flags().BoolVar(jsonOut, "json", "j", false, "JSON 형식으로 출력")
	return ask
}

func newAgentCommand(model, ollamaURL *string, jsonOut *bool) *wcli.Command {
	agent := &wcli.Command{
		Use:   "agent <이미지파일> <요청>",
		Short: "Ollama Vision 에이전트로 이미지 작업 분석",
		Long:  "이미지를 분석하고 필요한 경우 읽기 전용 이미지 도구를 호출한다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) < 2 {
				return errors.New("이미지 파일과 요청을 지정하세요 (예: tdraw agent photo.jpg \"크기와 내용을 알려줘\")")
			}
			configuredModel := *model
			if configuredModel == "" {
				configuredModel = envOrDefault("TDRAW_LLM_MODEL", "llava")
			}
			configuredURL := *ollamaURL
			if configuredURL == "" {
				configuredURL = envOrDefault("TDRAW_OLLAMA_URL", "http://localhost:11434/v1")
			}

			client := ollama.New(ollama.Config{BaseURL: configuredURL})
			request := strings.Join(ctx.Args[1:], " ")
			answer, err := vision.RunAgentFile(ctx.Context, client, configuredModel, ctx.Args[0], request)
			if err != nil {
				return err
			}
			if *jsonOut {
				fmt.Printf("{\"image\":%q,\"model\":%q,\"answer\":%q}\n", ctx.Args[0], configuredModel, answer)
				return nil
			}
			fmt.Println(answer)
			return nil
		},
	}
	agent.Flags().StringVar(model, "model", "m", "", "Ollama Vision 모델 (기본값: TDRAW_LLM_MODEL 또는 llava)")
	agent.Flags().StringVar(ollamaURL, "ollama-url", "", "", "Ollama OpenAI 호환 주소 (기본값: TDRAW_OLLAMA_URL 또는 localhost:11434/v1)")
	agent.Flags().BoolVar(jsonOut, "json", "j", false, "JSON 형식으로 출력")
	return agent
}

func newOCRCommand(model, ollamaURL *string, jsonOut *bool) *wcli.Command {
	ocr := &wcli.Command{
		Use:   "ocr <이미지파일>",
		Short: "Ollama Vision 모델로 이미지의 텍스트 추출",
		Long:  "이미지의 텍스트를 줄바꿈과 읽기 순서를 유지해 추출한다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 1 {
				return errors.New("이미지 파일을 하나 지정하세요 (예: tdraw ocr screenshot.png)")
			}
			configuredModel := *model
			if configuredModel == "" {
				configuredModel = envOrDefault("TDRAW_LLM_MODEL", "llava")
			}
			configuredURL := *ollamaURL
			if configuredURL == "" {
				configuredURL = envOrDefault("TDRAW_OLLAMA_URL", "http://localhost:11434/v1")
			}

			client := ollama.New(ollama.Config{BaseURL: configuredURL})
			text, err := vision.OCRFile(ctx.Context, client, configuredModel, ctx.Args[0])
			if err != nil {
				return err
			}
			if *jsonOut {
				fmt.Printf("{\"image\":%q,\"model\":%q,\"text\":%q}\n", ctx.Args[0], configuredModel, text)
				return nil
			}
			fmt.Println(text)
			return nil
		},
	}
	ocr.Flags().StringVar(model, "model", "m", "", "Ollama Vision 모델 (기본값: TDRAW_LLM_MODEL 또는 llava)")
	ocr.Flags().StringVar(ollamaURL, "ollama-url", "", "", "Ollama OpenAI 호환 주소 (기본값: TDRAW_OLLAMA_URL 또는 localhost:11434/v1)")
	ocr.Flags().BoolVar(jsonOut, "json", "j", false, "JSON 형식으로 출력")
	return ocr
}

func newIndexCommand(model, ollamaURL, output *string) *wcli.Command {
	indexCmd := &wcli.Command{
		Use:   "index <디렉터리>",
		Short: "이미지 설명 인덱스 생성",
		Long:  "디렉터리의 이미지를 Ollama Vision으로 분석해 로컬 JSON 인덱스를 생성한다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 1 {
				return errors.New("인덱싱할 디렉터리를 하나 지정하세요")
			}
			root := ctx.Args[0]
			indexPath := *output
			if indexPath == "" {
				indexPath = filepath.Join(root, ".tdraw", "index.json")
			}
			configuredModel := *model
			if configuredModel == "" {
				configuredModel = envOrDefault("TDRAW_LLM_MODEL", "llava")
			}
			configuredURL := *ollamaURL
			if configuredURL == "" {
				configuredURL = envOrDefault("TDRAW_OLLAMA_URL", "http://localhost:11434/v1")
			}
			client := ollama.New(ollama.Config{BaseURL: configuredURL})
			idx, err := imgindex.Build(ctx.Context, root, client, configuredModel, func(path string) {
				fmt.Fprintf(os.Stderr, "분석 중: %s\n", path)
			})
			if err != nil {
				return err
			}
			if err := idx.Save(indexPath); err != nil {
				return fmt.Errorf("인덱스 저장 실패: %w", err)
			}
			fmt.Fprintf(os.Stderr, "인덱스 생성 완료: %s (%d개)\n", indexPath, len(idx.Entries))
			return nil
		},
	}
	indexCmd.Flags().StringVar(model, "model", "m", "", "Ollama Vision 모델")
	indexCmd.Flags().StringVar(ollamaURL, "ollama-url", "", "", "Ollama OpenAI 호환 주소")
	indexCmd.Flags().StringVar(output, "output", "o", "", "인덱스 출력 경로 (기본값: <디렉터리>/.tdraw/index.json)")
	return indexCmd
}

func newFindCommand(indexPath *string, limit *int, jsonOut *bool) *wcli.Command {
	findCmd := &wcli.Command{
		Use:   "find <디렉터리> <검색어>",
		Short: "인덱스에서 이미지 검색",
		Long:  "로컬 이미지 인덱스의 경로와 Vision 설명을 대상으로 검색한다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) < 2 {
				return errors.New("검색할 디렉터리와 검색어를 지정하세요")
			}
			path := *indexPath
			if path == "" {
				path = filepath.Join(ctx.Args[0], ".tdraw", "index.json")
			}
			idx, err := imgindex.Load(path)
			if err != nil {
				return fmt.Errorf("인덱스 로드 실패: %w (먼저 tdraw index를 실행하세요)", err)
			}
			results := idx.Search(strings.Join(ctx.Args[1:], " "), *limit)
			if *jsonOut {
				data, err := json.Marshal(results)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			for _, entry := range results {
				fmt.Printf("%s\t%s\n", entry.Path, entry.Description)
			}
			return nil
		},
	}
	findCmd.Flags().StringVar(indexPath, "index", "i", "", "인덱스 파일 경로")
	findCmd.Flags().IntVar(limit, "limit", "n", 20, "최대 결과 수")
	findCmd.Flags().BoolVar(jsonOut, "json", "j", false, "JSON 형식으로 출력")
	return findCmd
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func run(ctx *wcli.Context, width int, colorMode string) error {
	args := ctx.Args
	if len(args) == 0 {
		rich.Fprintln(os.Stderr, "[red]이미지 파일을 지정하세요 (도움말: tdraw -h)[/red]")
		return errFailed
	}

	mode := parseColorMode(colorMode)

	cols, _ := render.TermSize()
	targetW := cols
	if width > 0 {
		targetW = width
	}
	targetH := targetW * heightScaleFactor

	logging.Debug("렌더링 시작: 파일 %d개, 출력폭=%d, 컬러=%s", len(args), targetW, colorMode)

	// Ctrl+C로 GIF 애니메이션 재생을 중단할 수 있도록 시그널 컨텍스트 사용
	sigCtx, stop := signal.NotifyContext(ctx.Context, os.Interrupt)
	defer stop()

	failed := false
	for i, path := range args {
		if i > 0 {
			fmt.Println()
		}
		if err := showImage(sigCtx, path, targetW, targetH, mode); err != nil {
			rich.Fprintln(os.Stderr, "[red]오류 (%s): %s[/red]", filepath.Base(path), err.Error())
			failed = true
		}
		if sigCtx.Err() != nil {
			break // Ctrl+C로 중단됨
		}
	}

	if failed {
		return errFailed
	}
	return nil
}

func showImage(ctx context.Context, path string, targetW, targetH int, mode render.ColorMode) error {
	// GIF는 애니메이션으로 재생
	if strings.ToLower(filepath.Ext(path)) == ".gif" {
		return showGIF(ctx, path, targetW, targetH, mode)
	}

	name := filepath.Base(path)

	// 진행 표시는 stderr로 출력해 stdout(이미지) 파이프를 오염시키지 않는다.
	sp := rich.NewSpinner(os.Stderr)
	sp.Start("이미지 로딩 중")
	t0 := time.Now()
	info, err := imgutil.Load(path)
	if err != nil {
		sp.Stop("")
		return err
	}
	logging.Debug("%s 디코드: %s %dx%d (%v)", name, info.Format, info.Width, info.Height, time.Since(t0))

	sp.UpdateText("리사이즈 중")
	t1 := time.Now()
	resized := imgutil.Resize(info.Image, targetW, targetH)
	sp.Stop("")

	rb := resized.Bounds()
	logging.Debug("%s 리사이즈: %dx%d (%v)", name, rb.Dx(), rb.Dy(), time.Since(t1))

	// /2: 반블록이므로 출력 높이는 실제 터미널 행 수의 절반
	printMetaBox("📄 "+name, "cyan",
		info.Format, info.Width, info.Height, rb.Dx(), rb.Dy()/2, -1)

	render.Render(os.Stdout, resized, mode)
	return nil
}

func showGIF(ctx context.Context, path string, targetW, targetH int, mode render.ColorMode) error {
	name := filepath.Base(path)

	sp := rich.NewSpinner(os.Stderr)
	sp.Start("GIF 디코딩 중")
	t0 := time.Now()
	anim, err := imgutil.LoadGIF(path)
	if err != nil {
		sp.Stop("")
		return err
	}
	sp.Stop("")
	logging.Debug("%s GIF 디코드: %dx%d, 프레임 %d개 (%v)", name, anim.Width, anim.Height, len(anim.Frames), time.Since(t0))

	// 모든 프레임을 동일한 크기로 리사이즈 (캔버스 크기가 같으므로 결과도 동일).
	// 프레임이 많아 체감 지연이 있으므로 진행률을 ProgressBar로 표시한다(stderr).
	frames := make([]image.Image, len(anim.Frames))
	pb := rich.NewProgressBar(len(anim.Frames))
	pb.Width = 24
	pb.FillColor = "magenta"
	pb.EmptyColor = "dim"
	pb.ShowCounter = true
	tty := isTTY(os.Stderr)
	t1 := time.Now()
	for i, f := range anim.Frames {
		frames[i] = imgutil.Resize(f, targetW, targetH)
		if tty {
			fmt.Fprint(os.Stderr, "\r"+rich.Sprint("[magenta]프레임 처리[/magenta] %s", pb.Render(i+1)))
		}
	}
	if tty {
		fmt.Fprint(os.Stderr, "\r\033[K") // 진행바가 있던 줄을 지운다
	}

	if len(frames) == 0 { // LoadGIF가 막지만 방어적으로 한 번 더 확인
		return fmt.Errorf("GIF에 표시할 프레임이 없습니다")
	}
	rb := frames[0].Bounds()
	logging.Debug("%s 프레임 리사이즈: %dx%d x%d개 (%v)", name, rb.Dx(), rb.Dy(), len(frames), time.Since(t1))

	printMetaBox("🎬 "+name, "magenta",
		"GIF", anim.Width, anim.Height, rb.Dx(), rb.Dy()/2, len(frames))

	render.PlayGIF(ctx, os.Stdout, frames, anim.Delays, mode)
	return nil
}

// printMetaBox는 이미지 메타정보를 rich.Box로 stdout에 출력한다.
// 라벨 세로 정렬은 rich.DisplayWidth(전각=2 반영)로 계산하고, 테두리 정렬은
// rich.Box가 처리한다. accent는 제목 색상, frames > 0이면 GIF로 간주한다.
func printMetaBox(title, accent, format string, origW, origH, outW, outH, frames int) {
	type kv struct{ k, v string }
	rows := []kv{
		{"포맷", strings.ToUpper(format)},
		{"원본", fmt.Sprintf("%d x %d", origW, origH)},
		{"출력", fmt.Sprintf("%d x %d", outW, outH)},
	}
	if frames > 0 {
		rows = append(rows, kv{"프레임", fmt.Sprintf("%d  (Ctrl+C로 종료)", frames)})
	}

	// 가장 넓은 라벨 기준으로 값을 세로 정렬한다.
	labelW := 0
	for _, r := range rows {
		if w := rich.DisplayWidth(r.k); w > labelW {
			labelW = w
		}
	}

	var b strings.Builder
	for i, r := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		gap := strings.Repeat(" ", labelW-rich.DisplayWidth(r.k)+2)
		fmt.Fprintf(&b, "[dim]%s[/dim]%s%s", r.k, gap, r.v)
	}

	rich.NewBox(b.String()).
		WithTitle(fmt.Sprintf("[%s]%s[/%s]", accent, title, accent)).
		Render(os.Stdout)
}

// setupLogger는 진단 로거를 stderr에 설정한다.
// verbose면 Debug, 아니면 Warn 이상만 출력해 평상시 조용하게 유지한다.
func setupLogger(verbose bool) {
	level := logging.LevelWarn
	if verbose {
		level = logging.LevelDebug
	}
	logging.SetLogger(logging.NewDefaultLogger(os.Stderr, level, true))
}

// isTTY는 f가 터미널에 연결되어 있는지 반환한다.
func isTTY(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

func parseColorMode(s string) render.ColorMode {
	switch strings.ToLower(s) {
	case "256":
		return render.Color256
	case "gray", "grey":
		return render.ColorGray
	default:
		return render.ColorTruecolor
	}
}

func validateColorMode(val string) error {
	switch strings.ToLower(val) {
	case "truecolor", "256", "gray", "grey":
		return nil
	default:
		return fmt.Errorf("컬러 모드는 truecolor | 256 | gray 중 하나여야 합니다 (입력값: %q)", val)
	}
}
