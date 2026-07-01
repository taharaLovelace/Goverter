package convert

import (
	"testing"

	"github.com/taharaLovelace/Goverter/internal/media"
)

func TestLookupFormatNormalizesAliases(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"jpg", "JPEG", ".jpeg"} {
		format, err := LookupFormat(input)
		if err != nil {
			t.Fatalf("LookupFormat(%q): %v", input, err)
		}
		if format.Name != "jpg" {
			t.Fatalf("LookupFormat(%q).Name = %q", input, format.Name)
		}
	}
}

func TestCompatibilityMatrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input, output media.Kind
		want          bool
	}{
		{media.KindVideo, media.KindVideo, true},
		{media.KindVideo, media.KindAudio, true},
		{media.KindAudio, media.KindAudio, true},
		{media.KindImage, media.KindImage, true},
		{media.KindAudio, media.KindVideo, false},
		{media.KindImage, media.KindVideo, false},
		{media.KindVideo, media.KindImage, false},
	}
	for _, test := range tests {
		if got := Compatible(test.input, test.output); got != test.want {
			t.Errorf("Compatible(%q, %q) = %v, want %v", test.input, test.output, got, test.want)
		}
	}
}

func TestParsePresetRejectsUnknownValue(t *testing.T) {
	t.Parallel()
	if _, err := ParsePreset("huge"); err == nil {
		t.Fatal("expected unknown preset to fail")
	}
}
