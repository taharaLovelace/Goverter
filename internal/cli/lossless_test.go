package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestExplicitLosslessPresetIsUsageErrorBeforeToolDiscovery(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := Execute(
		context.Background(),
		[]string{"convert", "input.wav", "--to", "flac", "--preset", "quality"},
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "does not apply") {
		t.Fatalf("unexpected error: %s", stderr.String())
	}
}
