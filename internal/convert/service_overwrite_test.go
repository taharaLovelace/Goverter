package convert

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/taharaLovelace/Goverter/internal/media"
)

func TestServiceAtomicallyReplacesExistingOutputWhenRequested(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "movie.mov")
	output := filepath.Join(directory, "movie.mp4")
	if err := os.WriteFile(input, []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("old media"), 0o644); err != nil {
		t.Fatal(err)
	}

	service := Service{
		Prober: fakeProber{items: map[string]media.Info{input: videoInfo()}},
		Runner: fakeRunner{},
	}
	summary, err := service.Convert(context.Background(), input, Options{
		Target: "mp4", Preset: "balanced", Overwrite: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Succeeded != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	data, readErr := os.ReadFile(output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "complete media" {
		t.Fatalf("output = %q, want replacement", data)
	}
	assertNoTemporaryFiles(t, directory)
}
