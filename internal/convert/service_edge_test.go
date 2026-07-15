package convert

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/taharaLovelace/Goverter/internal/media"
)

type countingRunner struct {
	calls int
}

func (r *countingRunner) Run(_ context.Context, plan Plan, update func(float64, bool)) error {
	r.calls++
	if r.calls == 1 {
		os.WriteFile(plan.Output, []byte("partial"), 0o644)
		return errors.New("first conversion failed")
	}
	if err := os.WriteFile(plan.Output, []byte("complete"), 0o644); err != nil {
		return err
	}
	update(100, true)
	return nil
}

func TestDirectoryContinuesAfterIndividualFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := filepath.Join(directory, "a.wav")
	second := filepath.Join(directory, "b.wav")
	os.WriteFile(first, []byte("input"), 0o644)
	os.WriteFile(second, []byte("input"), 0o644)

	runner := &countingRunner{}
	service := Service{
		Prober: fakeProber{items: map[string]media.Info{
			first:  audioInfo(),
			second: audioInfo(),
		}},
		Runner: runner,
	}
	summary, err := service.Convert(context.Background(), directory, Options{
		Target: "mp3", Preset: "balanced",
	}, nil)
	if err == nil {
		t.Fatal("expected a partial failure")
	}
	if runner.calls != 2 || summary.Failed != 1 || summary.Succeeded != 1 {
		t.Fatalf("batch did not continue: calls=%d summary=%#v", runner.calls, summary)
	}
	if _, statErr := os.Stat(filepath.Join(directory, "converted", "a.mp3")); !os.IsNotExist(statErr) {
		t.Fatalf("failed item left final output: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(directory, "converted", "b.mp3")); statErr != nil {
		t.Fatalf("second output missing: %v", statErr)
	}
}

type canceledRunner struct{}

func (canceledRunner) Run(_ context.Context, plan Plan, _ func(float64, bool)) error {
	os.WriteFile(plan.Output, []byte("partial"), 0o644)
	return context.Canceled
}

func TestCancellationRemovesTemporaryOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "movie.mov")
	os.WriteFile(input, []byte("input"), 0o644)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := Service{
		Prober: fakeProber{items: map[string]media.Info{input: videoInfo()}},
		Runner: canceledRunner{},
	}
	_, err := service.Convert(ctx, input, Options{
		Target: "mp4", Preset: "balanced",
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
	if _, statErr := os.Stat(filepath.Join(directory, "movie.mp4")); !os.IsNotExist(statErr) {
		t.Fatalf("canceled conversion left final output: %v", statErr)
	}
	assertNoTemporaryFiles(t, directory)
}

func TestRecursiveCollectionExcludesOutputSubtree(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	output := filepath.Join(directory, "converted")
	os.MkdirAll(output, 0o755)
	os.WriteFile(filepath.Join(directory, "source.wav"), []byte("input"), 0o644)
	os.WriteFile(filepath.Join(output, "old.wav"), []byte("output"), 0o644)

	files, err := collectFiles(directory, output, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "source.wav" {
		t.Fatalf("collected files = %#v", files)
	}
}

func TestSamePathUsesNativeCaseSensitivity(t *testing.T) {
	got := samePath(filepath.Join("root", "converted"), filepath.Join("root", "Converted"))
	want := runtime.GOOS == "windows"
	if got != want {
		t.Fatalf("samePath() = %t, want %t on %s", got, want, runtime.GOOS)
	}
}
