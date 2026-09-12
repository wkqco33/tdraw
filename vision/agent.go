package vision

import (
	"context"
	"encoding/json"
	"fmt"

	llm "github.com/wkqco33/LLM_client_go"
	llmagent "github.com/wkqco33/LLM_client_go/agent"
	"github.com/wkqco33/LLM_client_go/openai"
	"github.com/wkqco33/tdraw/imgutil"
)

// RunAgentFile runs a bounded, read-only image agent.
func RunAgentFile(ctx context.Context, client llm.Client, model, path, request string) (string, error) {
	message, err := FileMessage(path, request)
	if err != nil {
		return "", err
	}

	runner := llmagent.NewRunner(client, model,
		llmagent.WithMaxTurns(4),
		llmagent.WithSystemPrompt("You are tdraw, a terminal image assistant. Answer in the user's language. "+
			"Use image_metadata when exact file dimensions or format are needed. "+
			"Do not claim to modify or save files; this agent is read-only."),
	)
	runner.RegisterTool(&imageMetadataTool{path: path})

	_, resp, err := runner.Run(ctx, []llm.Message{message})
	if err != nil {
		return "", fmt.Errorf("에이전트 실행 실패: %w", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("에이전트 응답이 비어 있습니다")
	}
	return resp.Choices[0].Message.Content, nil
}

type imageMetadataTool struct {
	path string
}

func (t *imageMetadataTool) Definition() llm.Tool {
	return openai.NewTool(
		"image_metadata",
		"Get the exact format and dimensions of the current image.",
		map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	)
}

func (t *imageMetadataTool) Execute(_ context.Context, _ string) (string, error) {
	info, err := imgutil.LoadConfig(t.path)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{
		"path":   info.Path,
		"format": info.Format,
		"width":  info.Width,
		"height": info.Height,
	})
	return string(data), err
}
