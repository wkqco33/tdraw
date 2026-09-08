// Package vision provides image analysis through an LLM vision model.
package vision

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"

	llm "github.com/wkqco33/LLM_client_go"
	_ "github.com/wkqco33/tdraw/pnm"
)

// Completer is the part of an LLM client required for image questions.
// Keeping this small makes the image workflow easy to test and reuse.
type Completer interface {
	Complete(context.Context, llm.ChatRequest) (*llm.ChatResponse, error)
}

// FileMessage loads an image and creates a multimodal user message.
func FileMessage(path, question string) (llm.Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return llm.Message{}, fmt.Errorf("이미지 읽기 실패: %w", err)
	}
	if len(data) == 0 {
		return llm.Message{}, fmt.Errorf("빈 이미지 파일입니다")
	}

	mediaType := http.DetectContentType(data)
	if isPNM(data) {
		// Vision API 구현체는 PNM MIME 타입과 ASCII PNM 바이트를 지원하지
		// 않을 수 있으므로, API 호환성이 높은 PNG로 정규화한다.
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return llm.Message{}, fmt.Errorf("PNM 이미지 디코딩 실패: %w", err)
		}
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, img); err != nil {
			return llm.Message{}, fmt.Errorf("PNM 이미지 PNG 변환 실패: %w", err)
		}
		data = encoded.Bytes()
		mediaType = "image/png"
	} else if !strings.HasPrefix(mediaType, "image/") {
		return llm.Message{}, fmt.Errorf("이미지 파일이 아닙니다: %s", mediaType)
	}

	return llm.NewUserMessageWithParts(
		llm.TextContent(question),
		llm.ImageContentData(data, mediaType),
	), nil
}

// isPNM은 데이터가 Netpbm PNM 매직 넘버(P1-P6)로 시작하는지 확인한다.
func isPNM(data []byte) bool {
	if len(data) < 2 || data[0] != 'P' {
		return false
	}
	return data[1] >= '1' && data[1] <= '6'
}

// AskFile sends an image and question to a vision-capable model.
func AskFile(ctx context.Context, client Completer, model, path, question string) (string, error) {
	message, err := FileMessage(path, question)
	if err != nil {
		return "", err
	}

	resp, err := client.Complete(ctx, llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{message},
	})
	if err != nil {
		return "", fmt.Errorf("Vision 모델 요청 실패: %w", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("Vision 모델 응답이 비어 있습니다")
	}

	return resp.Choices[0].Message.Content, nil
}
