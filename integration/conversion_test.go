//go:build integration

package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	converter "github.com/taharaLovelace/Goverter/internal/convert"
	"github.com/taharaLovelace/Goverter/internal/media"
	"github.com/taharaLovelace/Goverter/internal/toolchain"
)

func TestRealFFmpegConversions(t *testing.T) {
	resolver := toolchain.NewResolver()
	ffmpeg, err := resolver.FFmpeg()
	if err != nil {
		t.Fatal(err)
	}
	ffprobe, err := resolver.FFprobe()
	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	video := filepath.Join(directory, "sample video.mp4")
	wav := filepath.Join(directory, "sample audio.wav")
	png := filepath.Join(directory, "sample image.png")

	runFFmpeg(t, ffmpeg,
		"-f", "lavfi", "-i", "testsrc=size=64x64:rate=10",
		"-f", "lavfi", "-i", "sine=frequency=1000:sample_rate=44100",
		"-t", "0.5", "-c:v", "libx264", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-y", video,
	)
	runFFmpeg(t, ffmpeg,
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100",
		"-t", "0.25", "-c:a", "pcm_s16le", "-y", wav,
	)
	runFFmpeg(t, ffmpeg,
		"-f", "lavfi", "-i", "color=c=blue:size=32x32",
		"-frames:v", "1", "-threads", "1", "-y", png,
	)

	service := converter.Service{
		Prober: media.FFprobe{Path: ffprobe},
		Runner: converter.FFmpegRunner{Path: ffmpeg},
	}
	cases := []struct {
		input  string
		target string
		output string
	}{
		{video, "webm", filepath.Join(directory, "sample video.webm")},
		{video, "mp3", filepath.Join(directory, "sample video.mp3")},
		{wav, "flac", filepath.Join(directory, "sample audio.flac")},
		{png, "webp", filepath.Join(directory, "sample image.webp")},
	}
	for _, test := range cases {
		t.Run(test.target, func(t *testing.T) {
			summary, convertErr := service.Convert(context.Background(), test.input, converter.Options{
				Target: test.target,
				Preset: "balanced",
				Output: test.output,
			}, nil)
			if convertErr != nil {
				t.Fatal(convertErr)
			}
			if summary.Succeeded != 1 {
				t.Fatalf("unexpected summary: %#v", summary)
			}
			info, probeErr := (media.FFprobe{Path: ffprobe}).Probe(context.Background(), test.output)
			if probeErr != nil {
				t.Fatal(probeErr)
			}
			if info.Kind == media.KindUnknown {
				t.Fatalf("output was not recognized: %#v", info)
			}
		})
	}
}

func runFFmpeg(t *testing.T, path string, args ...string) {
	t.Helper()
	command := exec.Command(path, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("create fixture with ffmpeg: %v", err)
	}
}
