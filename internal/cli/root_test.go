package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestFormatsJSON(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), []string{"formats", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	var output formatsOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Formats) != 9 || len(output.Presets) != 3 {
		t.Fatalf("unexpected formats output: %#v", output)
	}
}

func TestConvertUsageErrorsUseExitCodeTwo(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), []string{"convert", "input.mp4"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestUnknownCommandUsesExitCodeTwo(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), []string{"unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}
