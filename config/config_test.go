package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Width != 0 {
		t.Errorf("expected Width 0, got %d", cfg.Width)
	}
	if cfg.Color != "truecolor" {
		t.Errorf("expected Color 'truecolor', got %q", cfg.Color)
	}
	if cfg.NoMeta != false {
		t.Errorf("expected NoMeta false, got %v", cfg.NoMeta)
	}
	if cfg.Quiet != false {
		t.Errorf("expected Quiet false, got %v", cfg.Quiet)
	}
	if cfg.AI.Model != "llava" {
		t.Errorf("expected AI.Model 'llava', got %q", cfg.AI.Model)
	}
	if cfg.AI.OllamaURL != "http://localhost:11434/v1" {
		t.Errorf("expected AI.OllamaURL 'http://localhost:11434/v1', got %q", cfg.AI.OllamaURL)
	}
	if cfg.Find.Limit != 20 {
		t.Errorf("expected Find.Limit 20, got %d", cfg.Find.Limit)
	}
	if cfg.Play.Loop != false {
		t.Errorf("expected Play.Loop false, got %v", cfg.Play.Loop)
	}
}

func TestDefaultPath(t *testing.T) {
	customPath := filepath.Join(t.TempDir(), "custom_config.json")
	t.Setenv("TDRAW_CONFIG", customPath)
	p, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed: %v", err)
	}
	if p != customPath {
		t.Errorf("expected %q, got %q", customPath, p)
	}

	t.Setenv("TDRAW_CONFIG", "")
	p2, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed without TDRAW_CONFIG: %v", err)
	}
	if filepath.Base(p2) != "config.json" {
		t.Errorf("expected filename 'config.json', got %q", filepath.Base(p2))
	}
}

func TestLoadNonExistent(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "not_found.json")
	cfg, err := LoadFrom(nonExistent)
	if err != nil {
		t.Fatalf("expected nil error for non-existent file, got %v", err)
	}
	if cfg.Color != "truecolor" {
		t.Errorf("expected default config, got color %q", cfg.Color)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "config.json")

	cfg := DefaultConfig()
	cfg.Width = 100
	cfg.Color = "256"
	cfg.AI.Model = "custom-vision"
	cfg.Find.Limit = 50
	cfg.Play.Loop = true

	if err := cfg.SaveTo(p); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(p)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Width != 100 {
		t.Errorf("expected Width 100, got %d", loaded.Width)
	}
	if loaded.Color != "256" {
		t.Errorf("expected Color '256', got %q", loaded.Color)
	}
	if loaded.AI.Model != "custom-vision" {
		t.Errorf("expected AI.Model 'custom-vision', got %q", loaded.AI.Model)
	}
	if loaded.Find.Limit != 50 {
		t.Errorf("expected Find.Limit 50, got %d", loaded.Find.Limit)
	}
	if !loaded.Play.Loop {
		t.Errorf("expected Play.Loop true, got false")
	}
}

func TestDefaultSaveAndLoad(t *testing.T) {
	tempConfig := filepath.Join(t.TempDir(), "tdraw_config.json")
	t.Setenv("TDRAW_CONFIG", tempConfig)

	cfg := DefaultConfig()
	cfg.Width = 80
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Width != 80 {
		t.Errorf("expected Width 80, got %d", loaded.Width)
	}
}

func TestGetAndSet(t *testing.T) {
	cfg := DefaultConfig()

	// width
	if err := cfg.Set("width", "120"); err != nil {
		t.Fatalf("Set width failed: %v", err)
	}
	if val, err := cfg.Get("width"); err != nil || val != "120" {
		t.Errorf("Get width = %q, want '120'", val)
	}
	if err := cfg.Set("width", "-10"); err == nil {
		t.Errorf("expected error setting negative width")
	}
	if err := cfg.Set("width", "abc"); err == nil {
		t.Errorf("expected error setting non-integer width")
	}

	// color
	if err := cfg.Set("color", "gray"); err != nil {
		t.Fatalf("Set color failed: %v", err)
	}
	if val, err := cfg.Get("color"); err != nil || val != "gray" {
		t.Errorf("Get color = %q, want 'gray'", val)
	}
	if err := cfg.Set("color", "invalid"); err == nil {
		t.Errorf("expected error setting invalid color")
	}

	// no_meta & alias
	if err := cfg.Set("no-meta", "true"); err != nil {
		t.Fatalf("Set no-meta failed: %v", err)
	}
	if val, err := cfg.Get("no_meta"); err != nil || val != "true" {
		t.Errorf("Get no_meta = %q, want 'true'", val)
	}
	if err := cfg.Set("no_meta", "yes"); err != nil {
		t.Fatalf("Set no_meta with 'yes' failed: %v", err)
	}
	if val, err := cfg.Get("no_meta"); err != nil || val != "true" {
		t.Errorf("Get no_meta = %q, want 'true'", val)
	}

	// quiet
	if err := cfg.Set("quiet", "1"); err != nil {
		t.Fatalf("Set quiet failed: %v", err)
	}
	if val, err := cfg.Get("quiet"); err != nil || val != "true" {
		t.Errorf("Get quiet = %q, want 'true'", val)
	}

	// ai.model & aliases
	if err := cfg.Set("ai.model", "minicpm-v"); err != nil {
		t.Fatalf("Set ai.model failed: %v", err)
	}
	if val, err := cfg.Get("model"); err != nil || val != "minicpm-v" {
		t.Errorf("Get model = %q, want 'minicpm-v'", val)
	}
	if err := cfg.Set("model", ""); err == nil {
		t.Errorf("expected error setting empty model")
	}

	// ai.ollama_url & aliases
	if err := cfg.Set("ai.ollama_url", "http://ollama.local:11434/v1"); err != nil {
		t.Fatalf("Set ai.ollama_url failed: %v", err)
	}
	if val, err := cfg.Get("ollama_url"); err != nil || val != "http://ollama.local:11434/v1" {
		t.Errorf("Get ollama_url = %q, want 'http://ollama.local:11434/v1'", val)
	}
	if err := cfg.Set("ollama_url", "  "); err == nil {
		t.Errorf("expected error setting empty url")
	}

	// find.limit & aliases
	if err := cfg.Set("find.limit", "30"); err != nil {
		t.Fatalf("Set find.limit failed: %v", err)
	}
	if val, err := cfg.Get("limit"); err != nil || val != "30" {
		t.Errorf("Get limit = %q, want '30'", val)
	}
	if err := cfg.Set("find.limit", "0"); err == nil {
		t.Errorf("expected error setting limit 0")
	}

	// play.loop & aliases
	if err := cfg.Set("play.loop", "true"); err != nil {
		t.Fatalf("Set play.loop failed: %v", err)
	}
	if val, err := cfg.Get("loop"); err != nil || val != "true" {
		t.Errorf("Get loop = %q, want 'true'", val)
	}

	// unknown key
	if err := cfg.Set("unknown.key", "value"); err == nil {
		t.Errorf("expected error setting unknown key")
	}
	if _, err := cfg.Get("unknown.key"); err == nil {
		t.Errorf("expected error getting unknown key")
	}
}
