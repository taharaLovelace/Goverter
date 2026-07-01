package media

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Kind string

const (
	KindUnknown Kind = "unknown"
	KindVideo   Kind = "video"
	KindAudio   Kind = "audio"
	KindImage   Kind = "image"
)

type Stream struct {
	Index      int    `json:"index"`
	CodecType  string `json:"codec_type"`
	CodecName  string `json:"codec_name"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	SampleRate string `json:"sample_rate,omitempty"`
	Channels   int    `json:"channels,omitempty"`
	BitRate    string `json:"bit_rate,omitempty"`
	Duration   string `json:"duration,omitempty"`
	NBFrames   string `json:"nb_frames,omitempty"`
}

type Format struct {
	Filename   string            `json:"filename"`
	FormatName string            `json:"format_name"`
	Duration   string            `json:"duration"`
	Size       string            `json:"size"`
	BitRate    string            `json:"bit_rate"`
	Tags       map[string]string `json:"tags,omitempty"`
}

type ProbeData struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type StreamInfo struct {
	Index      int     `json:"index"`
	Type       string  `json:"type"`
	Codec      string  `json:"codec"`
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Channels   int     `json:"channels,omitempty"`
	BitRate    int64   `json:"bit_rate,omitempty"`
	Duration   float64 `json:"duration_seconds,omitempty"`
}

type Info struct {
	Path            string            `json:"path"`
	Kind            Kind              `json:"kind"`
	Format          string            `json:"format"`
	DurationSeconds float64           `json:"duration_seconds,omitempty"`
	SizeBytes       int64             `json:"size_bytes,omitempty"`
	BitRate         int64             `json:"bit_rate,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	Streams         []StreamInfo      `json:"streams"`
}

func ParseProbeJSON(data []byte, path string) (Info, error) {
	var raw ProbeData
	if err := json.Unmarshal(data, &raw); err != nil {
		return Info{}, fmt.Errorf("parse ffprobe output: %w", err)
	}
	if len(raw.Streams) == 0 {
		return Info{}, fmt.Errorf("no media streams found")
	}

	info := Info{
		Path:            path,
		Kind:            Classify(path, raw),
		Format:          raw.Format.FormatName,
		DurationSeconds: parseFloat(raw.Format.Duration),
		SizeBytes:       parseInt64(raw.Format.Size),
		BitRate:         parseInt64(raw.Format.BitRate),
		Tags:            raw.Format.Tags,
		Streams:         make([]StreamInfo, 0, len(raw.Streams)),
	}
	for _, stream := range raw.Streams {
		info.Streams = append(info.Streams, StreamInfo{
			Index:      stream.Index,
			Type:       stream.CodecType,
			Codec:      stream.CodecName,
			Width:      stream.Width,
			Height:     stream.Height,
			SampleRate: int(parseInt64(stream.SampleRate)),
			Channels:   stream.Channels,
			BitRate:    parseInt64(stream.BitRate),
			Duration:   parseFloat(stream.Duration),
		})
	}
	return info, nil
}

func (i Info) HasStream(streamType string) bool {
	for _, stream := range i.Streams {
		if stream.Type == streamType {
			return true
		}
	}
	return false
}

func parseFloat(value string) float64 {
	number, _ := strconv.ParseFloat(value, 64)
	return number
}

func parseInt64(value string) int64 {
	number, _ := strconv.ParseInt(value, 10, 64)
	return number
}

func primaryFormat(value string) string {
	if before, _, ok := strings.Cut(value, ","); ok {
		return before
	}
	return value
}
