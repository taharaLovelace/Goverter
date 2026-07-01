package media

import (
	"testing"
)

func TestParseProbeJSONClassifiesMedia(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		data string
		want Kind
	}{
		{
			name: "video with audio",
			path: "movie.mov",
			data: `{"streams":[{"index":0,"codec_type":"video","codec_name":"h264"},{"index":1,"codec_type":"audio","codec_name":"aac"}],"format":{"format_name":"mov,mp4","duration":"1.5","size":"2048"}}`,
			want: KindVideo,
		},
		{
			name: "audio",
			path: "sound.wav",
			data: `{"streams":[{"index":0,"codec_type":"audio","codec_name":"pcm_s16le","sample_rate":"48000","channels":2}],"format":{"format_name":"wav","duration":"2"}}`,
			want: KindAudio,
		},
		{
			name: "image",
			path: "photo.png",
			data: `{"streams":[{"index":0,"codec_type":"video","codec_name":"png","width":100,"height":50}],"format":{"format_name":"png_pipe"}}`,
			want: KindImage,
		},
		{
			name: "gif excluded",
			path: "animation.gif",
			data: `{"streams":[{"index":0,"codec_type":"video","codec_name":"gif"}],"format":{"format_name":"gif"}}`,
			want: KindUnknown,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			info, err := ParseProbeJSON([]byte(test.data), test.path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Kind != test.want {
				t.Fatalf("kind = %q, want %q", info.Kind, test.want)
			}
		})
	}
}

func TestParseProbeJSONRejectsInvalidData(t *testing.T) {
	t.Parallel()
	if _, err := ParseProbeJSON([]byte(`{"streams":[]}`), "empty.bin"); err == nil {
		t.Fatal("expected an error for input without streams")
	}
	if _, err := ParseProbeJSON([]byte(`not json`), "bad.bin"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}
