//go:build !tcamviewer || !cgo

package tcam

import (
	"errors"
	"testing"
)

func TestStubOpenReturnsErrUnsupported(t *testing.T) {
	src, err := Open("test.mp4", false)
	if src != nil {
		t.Fatalf("expected nil src from stub Open, got %v", src)
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}
