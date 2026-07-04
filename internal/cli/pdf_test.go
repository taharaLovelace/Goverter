package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPDFImagesRequiresInputAndOutput(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"pdf", "images"},
		{"pdf", "images", "photo.jpg"},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer
		code := Execute(context.Background(), args, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("Execute(%v) code = %d, want 2; stderr=%s", args, code, stderr.String())
		}
	}
}

func TestPDFImagesRejectsInvalidLayoutBeforeReadingInputs(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(
		context.Background(),
		[]string{
			"pdf", "images", "missing.jpg",
			"--output", "album.pdf",
			"--page-size", "legal",
		},
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unsupported page size") {
		t.Fatalf("unexpected error: %s", stderr.String())
	}
}

func TestPDFMergeRequiresTwoInputsAndOutput(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"pdf", "merge"},
		{"pdf", "merge", "first.pdf"},
		{"pdf", "merge", "first.pdf", "second.pdf"},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer
		code := Execute(context.Background(), args, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("Execute(%v) code = %d, want 2; stderr=%s", args, code, stderr.String())
		}
	}
}

func TestPDFMergeRejectsInvalidOutputBeforeReadingInputs(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(
		context.Background(),
		[]string{
			"pdf", "merge", "missing-one.pdf", "missing-two.pdf",
			"--output", "combined.txt",
		},
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), ".pdf extension") {
		t.Fatalf("unexpected error: %s", stderr.String())
	}
}
