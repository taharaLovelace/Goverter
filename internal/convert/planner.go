package convert

import (
	"fmt"

	"github.com/taharaLovelace/Goverter/internal/media"
)

type Plan struct {
	Input    string
	Output   string
	Duration float64
	Args     []string
}

func BuildPlan(input media.Info, output string, format Format, preset Preset) (Plan, error) {
	if !Compatible(input.Kind, format.Kind) {
		return Plan{}, fmt.Errorf("cannot convert %s input to %s output", input.Kind, format.Kind)
	}
	if format.Kind == media.KindAudio && !input.HasStream("audio") {
		return Plan{}, fmt.Errorf("input has no audio stream")
	}
	if (format.Kind == media.KindVideo || format.Kind == media.KindImage) && !input.HasStream("video") {
		return Plan{}, fmt.Errorf("input has no video stream")
	}

	args := []string{
		"-hide_banner",
		"-nostdin",
		"-nostats",
		"-progress", "pipe:1",
		"-i", input.Path,
		"-map_metadata", "0",
	}

	switch format.Kind {
	case media.KindVideo:
		args = append(args, "-map", "0:v:0", "-map", "0:a:0?")
	case media.KindAudio:
		args = append(args, "-map", "0:a:0", "-vn")
	case media.KindImage:
		args = append(args, "-map", "0:v:0", "-frames:v", "1", "-an")
	}

	args = append(args, codecArgs(format.Name, preset)...)
	if format.Kind == media.KindImage {
		args = append(args, "-update", "1")
	}
	args = append(args, "-f", format.Muxer, "-y", output)

	return Plan{
		Input:    input.Path,
		Output:   output,
		Duration: input.DurationSeconds,
		Args:     args,
	}, nil
}

func codecArgs(format string, preset Preset) []string {
	switch format {
	case "mp4":
		crf := presetValue(preset, "30", "23", "18")
		audioRate := presetValue(preset, "96k", "128k", "192k")
		return []string{
			"-c:v", "libx264", "-preset", "medium", "-crf", crf, "-pix_fmt", "yuv420p",
			"-c:a", "aac", "-b:a", audioRate, "-movflags", "+faststart",
		}
	case "webm":
		crf := presetValue(preset, "40", "32", "24")
		audioRate := presetValue(preset, "64k", "96k", "160k")
		return []string{
			"-c:v", "libvpx-vp9", "-crf", crf, "-b:v", "0",
			"-deadline", "good", "-cpu-used", "2",
			"-c:a", "libopus", "-b:a", audioRate,
		}
	case "mp3":
		return []string{"-c:a", "libmp3lame", "-q:a", presetValue(preset, "7", "4", "2")}
	case "wav":
		return []string{"-c:a", "pcm_s16le"}
	case "flac":
		return []string{"-c:a", "flac", "-compression_level", "5"}
	case "ogg":
		return []string{"-c:a", "libvorbis", "-q:a", presetValue(preset, "2", "5", "8")}
	case "jpg":
		return []string{"-c:v", "mjpeg", "-q:v", presetValue(preset, "8", "4", "2")}
	case "png":
		return []string{"-c:v", "png", "-compression_level", "6"}
	case "webp":
		return []string{"-c:v", "libwebp", "-quality", presetValue(preset, "60", "80", "95")}
	default:
		panic("unknown output format: " + format)
	}
}

func presetValue(preset Preset, compact, balanced, quality string) string {
	switch preset {
	case PresetCompact:
		return compact
	case PresetQuality:
		return quality
	default:
		return balanced
	}
}
