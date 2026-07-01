package convert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/taharaLovelace/Goverter/internal/media"
)

type Preset string

const (
	PresetCompact  Preset = "compact"
	PresetBalanced Preset = "balanced"
	PresetQuality  Preset = "quality"
)

var Presets = []Preset{PresetCompact, PresetBalanced, PresetQuality}

type Format struct {
	Name     string     `json:"name"`
	Aliases  []string   `json:"aliases,omitempty"`
	Kind     media.Kind `json:"kind"`
	Codecs   string     `json:"codecs"`
	Lossless bool       `json:"lossless"`
	Muxer    string     `json:"-"`
}

var formats = map[string]Format{
	"mp4": {
		Name: "mp4", Kind: media.KindVideo, Codecs: "H.264 / AAC", Muxer: "mp4",
	},
	"webm": {
		Name: "webm", Kind: media.KindVideo, Codecs: "VP9 / Opus", Muxer: "webm",
	},
	"mp3": {
		Name: "mp3", Kind: media.KindAudio, Codecs: "MP3 (LAME)", Muxer: "mp3",
	},
	"wav": {
		Name: "wav", Kind: media.KindAudio, Codecs: "PCM 16-bit", Lossless: true, Muxer: "wav",
	},
	"flac": {
		Name: "flac", Kind: media.KindAudio, Codecs: "FLAC", Lossless: true, Muxer: "flac",
	},
	"ogg": {
		Name: "ogg", Kind: media.KindAudio, Codecs: "Vorbis", Muxer: "ogg",
	},
	"jpg": {
		Name: "jpg", Aliases: []string{"jpeg"}, Kind: media.KindImage, Codecs: "Motion JPEG", Muxer: "image2",
	},
	"png": {
		Name: "png", Kind: media.KindImage, Codecs: "PNG", Lossless: true, Muxer: "image2",
	},
	"webp": {
		Name: "webp", Kind: media.KindImage, Codecs: "WebP", Muxer: "image2",
	},
}

func LookupFormat(name string) (Format, error) {
	normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), "."))
	if normalized == "jpeg" {
		normalized = "jpg"
	}
	format, ok := formats[normalized]
	if !ok {
		return Format{}, fmt.Errorf("unsupported output format %q; run \"goverter formats\" to list supported formats", name)
	}
	return format, nil
}

func ListFormats() []Format {
	result := make([]Format, 0, len(formats))
	for _, format := range formats {
		result = append(result, format)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return result[i].Name < result[j].Name
		}
		return result[i].Kind < result[j].Kind
	})
	return result
}

func ParsePreset(value string) (Preset, error) {
	preset := Preset(strings.ToLower(strings.TrimSpace(value)))
	for _, candidate := range Presets {
		if preset == candidate {
			return preset, nil
		}
	}
	return "", fmt.Errorf("unsupported preset %q; use compact, balanced, or quality", value)
}

func Compatible(input media.Kind, output media.Kind) bool {
	if input == output {
		return true
	}
	return input == media.KindVideo && output == media.KindAudio
}
