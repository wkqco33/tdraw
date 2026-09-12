// Package config는 tdraw의 전역 설정을 로드, 저장 및 관리하는 기능을 제공한다.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config는 tdraw의 전역 설정을 나타낸다.
type Config struct {
	Width  int        `json:"width"`
	Color  string     `json:"color"`
	NoMeta bool       `json:"no_meta"`
	Quiet  bool       `json:"quiet"`
	AI     AIConfig   `json:"ai"`
	Find   FindConfig `json:"find"`
	Play   PlayConfig `json:"play"`
}

// AIConfig는 Ollama Vision 등 AI 기능 관련 설정을 나타낸다.
type AIConfig struct {
	Model     string `json:"model"`
	OllamaURL string `json:"ollama_url"`
}

// FindConfig는 로컬 이미지 검색 관련 설정을 나타낸다.
type FindConfig struct {
	Limit int `json:"limit"`
}

// PlayConfig는 비디오 재생 관련 설정을 나타낸다.
type PlayConfig struct {
	Loop bool `json:"loop"`
}

// DefaultConfig는 권장 기본 설정값을 반환한다.
func DefaultConfig() Config {
	return Config{
		Width:  0,
		Color:  "truecolor",
		NoMeta: false,
		Quiet:  false,
		AI: AIConfig{
			Model:     "llava",
			OllamaURL: "http://localhost:11434/v1",
		},
		Find: FindConfig{
			Limit: 20,
		},
		Play: PlayConfig{
			Loop: false,
		},
	}
}

// DefaultPath는 현재 OS에 적합한 기본 설정 파일 경로를 반환한다.
// 환경변수 TDRAW_CONFIG가 지정되어 있으면 최우선으로 적용한다.
func DefaultPath() (string, error) {
	if p := os.Getenv("TDRAW_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("사용자 설정 디렉터리 확인 실패: %w", err)
	}
	return filepath.Join(dir, "tdraw", "config.json"), nil
}

// Load는 기본 경로에서 설정을 로드한다. 파일이 없으면 DefaultConfig를 반환한다.
func Load() (Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return DefaultConfig(), err
	}
	return LoadFrom(path)
}

// LoadFrom은 지정된 경로에서 설정을 로드한다.
// 파일이 존재하지 않으면 DefaultConfig를 에러 없이 반환한다.
func LoadFrom(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}
	return cfg, nil
}

// Save는 기본 경로에 현재 설정을 저장한다.
func (c Config) Save() error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return c.SaveTo(path)
}

// SaveTo는 지정된 경로에 설정을 JSON 형태로 저장한다.
// 상위 디렉터리가 없으면 자동으로 생성한다.
func (c Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("설정 디렉터리 생성 실패: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 쓰기 실패: %w", err)
	}
	return nil
}

// Get은 지정된 키의 설정값을 문자열로 반환한다.
func (c Config) Get(key string) (string, error) {
	switch normalizeKey(key) {
	case "width":
		return strconv.Itoa(c.Width), nil
	case "color":
		return c.Color, nil
	case "no_meta", "nometa":
		return strconv.FormatBool(c.NoMeta), nil
	case "quiet":
		return strconv.FormatBool(c.Quiet), nil
	case "ai.model", "ai_model", "model":
		return c.AI.Model, nil
	case "ai.ollama_url", "ai_ollama_url", "ollama_url", "ai.url", "ai_url":
		return c.AI.OllamaURL, nil
	case "find.limit", "find_limit", "limit":
		return strconv.Itoa(c.Find.Limit), nil
	case "play.loop", "play_loop", "loop":
		return strconv.FormatBool(c.Play.Loop), nil
	default:
		return "", fmt.Errorf("알 수 없는 설정 키: %q", key)
	}
}

// Set은 지정된 키의 설정값을 검증 후 갱신한다.
func (c *Config) Set(key, val string) error {
	switch normalizeKey(key) {
	case "width":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			return fmt.Errorf("width는 0 이상의 정수여야 합니다 (입력값: %q)", val)
		}
		c.Width = n
	case "color":
		v := strings.ToLower(val)
		if v != "truecolor" && v != "256" && v != "gray" && v != "grey" {
			return fmt.Errorf("color는 truecolor | 256 | gray 중 하나여야 합니다 (입력값: %q)", val)
		}
		c.Color = v
	case "no_meta", "nometa":
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("no_meta는 true 또는 false여야 합니다 (입력값: %q)", val)
		}
		c.NoMeta = b
	case "quiet":
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("quiet는 true 또는 false여야 합니다 (입력값: %q)", val)
		}
		c.Quiet = b
	case "ai.model", "ai_model", "model":
		v := strings.TrimSpace(val)
		if v == "" {
			return fmt.Errorf("ai.model은 비어 있을 수 없습니다")
		}
		c.AI.Model = v
	case "ai.ollama_url", "ai_ollama_url", "ollama_url", "ai.url", "ai_url":
		v := strings.TrimSpace(val)
		if v == "" {
			return fmt.Errorf("ai.ollama_url은 비어 있을 수 없습니다")
		}
		c.AI.OllamaURL = v
	case "find.limit", "find_limit", "limit":
		n, err := strconv.Atoi(val)
		if err != nil || n <= 0 {
			return fmt.Errorf("find.limit은 1 이상의 정수여야 합니다 (입력값: %q)", val)
		}
		c.Find.Limit = n
	case "play.loop", "play_loop", "loop":
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("play.loop는 true 또는 false여야 합니다 (입력값: %q)", val)
		}
		c.Play.Loop = b
	default:
		return fmt.Errorf("알 수 없는 설정 키: %q", key)
	}
	return nil
}

func normalizeKey(key string) string {
	s := strings.ToLower(strings.TrimSpace(key))
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "t", "yes", "y", "on":
		return true, nil
	case "false", "0", "f", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("잘못된 불리언 값: %q", s)
	}
}
