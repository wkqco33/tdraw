package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/wkqco33/tdraw/render"
)

func TestParseColorMode(t *testing.T) {
	tests := []struct {
		input string
		want  render.ColorMode
	}{
		{"256", render.Color256},
		{"gray", render.ColorGray},
		{"grey", render.ColorGray},
		{"truecolor", render.ColorTruecolor},
		{"unknown", render.ColorTruecolor},
	}

	for _, tt := range tests {
		got := parseColorMode(tt.input)
		if got != tt.want {
			t.Errorf("parseColorMode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateColorMode(t *testing.T) {
	valid := []string{"truecolor", "TRUECOLOR", "256", "gray", "GRAY", "grey"}
	for _, v := range valid {
		if err := validateColorMode(v); err != nil {
			t.Errorf("validateColorMode(%q) unexpectedly returned error: %v", v, err)
		}
	}

	invalid := []string{"invalid", "16", "cmyk"}
	for _, inv := range invalid {
		if err := validateColorMode(inv); err == nil {
			t.Errorf("validateColorMode(%q) expected error, got nil", inv)
		}
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("TDRAW_TEST_VAR", "configured_value")
	if got := envOrDefault("TDRAW_TEST_VAR", "fallback"); got != "configured_value" {
		t.Errorf("expected configured_value, got %s", got)
	}

	if got := envOrDefault("TDRAW_NONEXISTENT_VAR", "fallback"); got != "fallback" {
		t.Errorf("expected fallback, got %s", got)
	}
}

func TestSubcommandUsageErrors(t *testing.T) {
	// Ask without args
	var model, url string
	var jsonOut bool
	askCmd := newAskCommand(&model, &url, &jsonOut)
	err := askCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for ask without args, got %v", err)
	}

	// Agent without args
	agentCmd := newAgentCommand(&model, &url, &jsonOut)
	err = agentCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for agent without args, got %v", err)
	}

	// OCR without args
	ocrCmd := newOCRCommand(&model, &url, &jsonOut)
	err = ocrCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for ocr without args, got %v", err)
	}

	// Index without args
	var out string
	indexCmd := newIndexCommand(&model, &url, &out)
	err = indexCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for index without args, got %v", err)
	}

	// Find without args
	var limit int
	findCmd := newFindCommand(&out, &limit, &jsonOut)
	err = findCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for find without args, got %v", err)
	}

	// Play without args
	var loop bool
	var width int
	var color string
	playCmd := newPlayCommand(&loop, &width, &color)
	err = playCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for play without args, got %v", err)
	}

	// Config without args
	configCmd := newConfigCommand()
	err = configCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for config without args, got %v", err)
	}

	// Config get without args
	getCmd := newConfigGetCommand()
	err = getCmd.Execute([]string{})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for config get without args, got %v", err)
	}

	// Config set with insufficient args
	setCmd := newConfigSetCommand()
	err = setCmd.Execute([]string{"only_one_arg"})
	if err == nil || !errors.Is(err, errUsage) {
		t.Errorf("expected errUsage for config set with 1 arg, got %v", err)
	}
}

func TestConfigSubcommands(t *testing.T) {
	tempConfig := filepath.Join(t.TempDir(), "tdraw_test_cfg.json")
	t.Setenv("TDRAW_CONFIG", tempConfig)

	// 1. path
	pathCmd := newConfigPathCommand()
	if err := pathCmd.Execute([]string{}); err != nil {
		t.Fatalf("config path failed: %v", err)
	}

	// 2. init
	initCmd := newConfigInitCommand()
	if err := initCmd.Execute([]string{}); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// 3. init again without force (should fail with errFailed)
	initCmd2 := newConfigInitCommand()
	if err := initCmd2.Execute([]string{}); err == nil || !errors.Is(err, errFailed) {
		t.Fatalf("expected errFailed for init duplicate without force, got %v", err)
	}

	// 4. init with force
	initCmd3 := newConfigInitCommand()
	if err := initCmd3.Execute([]string{"-f"}); err != nil {
		t.Fatalf("config init -f failed: %v", err)
	}

	// 5. show
	showCmd := newConfigShowCommand()
	if err := showCmd.Execute([]string{}); err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	// 6. set valid key
	setCmd := newConfigSetCommand()
	if err := setCmd.Execute([]string{"ai.model", "my-vision-model"}); err != nil {
		t.Fatalf("config set ai.model failed: %v", err)
	}

	// 7. get valid key
	getCmd := newConfigGetCommand()
	if err := getCmd.Execute([]string{"ai.model"}); err != nil {
		t.Fatalf("config get ai.model failed: %v", err)
	}

	// 8. set invalid key
	setCmd2 := newConfigSetCommand()
	if err := setCmd2.Execute([]string{"nonexistent.key", "value"}); err == nil || !errors.Is(err, errFailed) {
		t.Fatalf("expected errFailed for invalid key set, got %v", err)
	}

	// 9. get invalid key
	getCmd2 := newConfigGetCommand()
	if err := getCmd2.Execute([]string{"nonexistent.key"}); err == nil || !errors.Is(err, errFailed) {
		t.Fatalf("expected errFailed for invalid key get, got %v", err)
	}
}
