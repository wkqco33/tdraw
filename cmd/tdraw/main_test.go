package main

import (
	"errors"
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
}
