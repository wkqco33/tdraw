package vision

import "context"

const ocrPrompt = "이미지에 포함된 모든 텍스트를 OCR해줘. 원래 줄바꿈과 읽기 순서를 유지하고, " +
	"텍스트가 없으면 없다고만 답해줘. 추측하거나 설명을 추가하지 마."

// OCRFile extracts visible text from an image through a vision-capable model.
func OCRFile(ctx context.Context, client Completer, model, path string) (string, error) {
	return AskFile(ctx, client, model, path, ocrPrompt)
}
