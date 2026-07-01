package convert

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/taharaLovelace/Goverter/internal/media"
)

type fakeProber struct {
	items map[string]media.Info
}

func (p fakeProber) Probe(_ context.Context, path string) (media.Info, error) {
	item, ok := p.items[path]
	if !ok {
		return media.Info{}, errors.New("not media")
	}
	item.Path = path
	return item, nil
}

type fakeRunner struct {
	fail bool
}

func (r fakeRunner) Run(_ context.Context, plan Plan, update func(float64, bool)) error {
	if err := os.WriteFile(plan.Output, []byte("complete media"), 0o644); err != nil {
		return err
	}
	update(100, true)
	if r.fail {
		return errors.New("encoder failed")
	}
	return nil
}

func TestServiceConvertsSingleFileWithoutLeavingTemporaryOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "vídeo source.mov")
	if err := os.WriteFile(input, []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}
	service := Service{
		Prober: fakeProber{items: map[string]media.Info{
			input: videoInfo(),
		}},
		Runner: fakeRunner{},
	}
	summary, err := service.Convert(context.Background(), input, Options{
		Target: "mp4", Preset: "balanced",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "vídeo source.mp4")
	if summary.Succeeded != 1 || summary.Results[0].Output != output {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "complete media" {
		t.Fatalf("output = %q", data)
	}
	assertNoTemporaryFiles(t, directory)
}

func TestServicePreservesExistingOutputByDefault(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "movie.mov")
	output := filepath.Join(directory, "movie.mp4")
	os.WriteFile(input, []byte("input"), 0o644)
	os.WriteFile(output, []byte("keep me"), 0o644)

	service := Service{
		Prober: fakeProber{items: map[string]media.Info{input: videoInfo()}},
		Runner: fakeRunner{},
	}
	summary, err := service.Convert(context.Background(), input, Options{
		Target: "mp4", Preset: "balanced",
	}, nil)
	if err == nil || summary.Failed != 1 {
		t.Fatalf("expected one failed conversion, got summary=%#v err=%v", summary, err)
	}
	data, _ := os.ReadFile(output)
	if string(data) != "keep me" {
		t.Fatalf("existing output was modified: %q", data)
	}
}

func TestServiceRemovesPartialOutputAfterFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "movie.mov")
	os.WriteFile(input, []byte("input"), 0o644)

	service := Service{
		Prober: fakeProber{items: map[string]media.Info{input: videoInfo()}},
		Runner: fakeRunner{fail: true},
	}
	summary, err := service.Convert(context.Background(), input, Options{
		Target: "mp4", Preset: "balanced",
	}, nil)
	if err == nil || summary.Failed != 1 {
		t.Fatalf("expected failure, got summary=%#v err=%v", summary, err)
	}
	if _, statErr := os.Stat(filepath.Join(directory, "movie.mp4")); !os.IsNotExist(statErr) {
		t.Fatalf("final output exists after failed conversion: %v", statErr)
	}
	assertNoTemporaryFiles(t, directory)
}

func TestServiceConvertsRecursiveDirectoryAndSkipsIncompatibleFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	nested := filepath.Join(directory, "nested")
	os.MkdirAll(nested, 0o755)
	video := filepath.Join(directory, "movie.mov")
	audio := filepath.Join(nested, "sound.wav")
	text := filepath.Join(directory, "notes.txt")
	for _, path := range []string{video, audio, text} {
		os.WriteFile(path, []byte("input"), 0o644)
	}

	service := Service{
		Prober: fakeProber{items: map[string]media.Info{
			video: videoInfo(),
			audio: audioInfo(),
		}},
		Runner: fakeRunner{},
	}
	summary, err := service.Convert(context.Background(), directory, Options{
		Target: "mp3", Preset: "balanced", Recursive: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 3 || summary.Succeeded != 2 || summary.Skipped != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	for _, output := range []string{
		filepath.Join(directory, "converted", "movie.mp3"),
		filepath.Join(directory, "converted", "nested", "sound.mp3"),
	} {
		if _, statErr := os.Stat(output); statErr != nil {
			t.Errorf("expected output %s: %v", output, statErr)
		}
	}
}

func TestServiceRejectsExplicitPresetForLosslessOutput(t *testing.T) {
	t.Parallel()
	service := Service{}
	_, err := service.Convert(context.Background(), "unused", Options{
		Target: "flac", Preset: "quality", PresetExplicit: true,
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "does not apply") {
		t.Fatalf("expected lossless preset error, got %v", err)
	}
}

func videoInfo() media.Info {
	return media.Info{
		Kind:            media.KindVideo,
		DurationSeconds: 1,
		Streams: []media.StreamInfo{
			{Type: "video"},
			{Type: "audio"},
		},
	}
}

func audioInfo() media.Info {
	return media.Info{
		Kind:            media.KindAudio,
		DurationSeconds: 1,
		Streams:         []media.StreamInfo{{Type: "audio"}},
	}
}

func assertNoTemporaryFiles(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".goverter-") {
			t.Errorf("temporary file was not removed: %s", entry.Name())
		}
	}
}
