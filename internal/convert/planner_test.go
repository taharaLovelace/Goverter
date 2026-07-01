package convert

import (
	"slices"
	"testing"

	"github.com/taharaLovelace/Goverter/internal/media"
)

func TestBuildPlanUsesSeparateArguments(t *testing.T) {
	t.Parallel()
	format, _ := LookupFormat("mp4")
	input := media.Info{
		Path:            `C:\Media Files\vídeo.mov`,
		Kind:            media.KindVideo,
		DurationSeconds: 2,
		Streams: []media.StreamInfo{
			{Type: "video"},
			{Type: "audio"},
		},
	}
	plan, err := BuildPlan(input, `C:\Output Files\vídeo.tmp`, format, PresetBalanced)
	if err != nil {
		t.Fatal(err)
	}
	index := slices.Index(plan.Args, "-i")
	if index < 0 || plan.Args[index+1] != input.Path {
		t.Fatalf("input path was not preserved as one argument: %#v", plan.Args)
	}
	if !slices.Contains(plan.Args, "libx264") || !slices.Contains(plan.Args, "+faststart") {
		t.Fatalf("MP4 codec arguments missing: %#v", plan.Args)
	}
}

func TestBuildPlanExtractsFirstAudioStream(t *testing.T) {
	t.Parallel()
	format, _ := LookupFormat("mp3")
	input := media.Info{
		Path: "video.mp4",
		Kind: media.KindVideo,
		Streams: []media.StreamInfo{
			{Type: "video"},
			{Type: "audio"},
		},
	}
	plan, err := BuildPlan(input, "audio.tmp", format, PresetCompact)
	if err != nil {
		t.Fatal(err)
	}
	if !containsPair(plan.Args, "-map", "0:a:0") || !slices.Contains(plan.Args, "-vn") {
		t.Fatalf("audio mapping missing: %#v", plan.Args)
	}
}

func TestBuildPlanRejectsInvalidCrossMediaConversion(t *testing.T) {
	t.Parallel()
	format, _ := LookupFormat("mp4")
	_, err := BuildPlan(media.Info{Kind: media.KindAudio}, "out.tmp", format, PresetBalanced)
	if err == nil {
		t.Fatal("expected audio-to-video conversion to fail")
	}
}

func containsPair(values []string, first, second string) bool {
	for index := 0; index+1 < len(values); index++ {
		if values[index] == first && values[index+1] == second {
			return true
		}
	}
	return false
}
