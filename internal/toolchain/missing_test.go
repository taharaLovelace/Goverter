package toolchain

import (
	"errors"
	"strings"
	"testing"
)

func TestResolverExplainsMissingTools(t *testing.T) {
	t.Parallel()
	resolver := Resolver{
		ExecutablePath: func() (string, error) { return "C:/Goverter/goverter.exe", nil },
		LookupPath:     func(string) (string, error) { return "", errors.New("missing") },
		Getenv:         func(string) string { return "" },
	}
	_, err := resolver.FFprobe()
	if err == nil || !strings.Contains(err.Error(), envDirectory) {
		t.Fatalf("unexpected error: %v", err)
	}
}
