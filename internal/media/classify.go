package media

import (
	"path/filepath"
	"strings"
)

var imageExtensions = map[string]bool{
	".bmp": true, ".jpeg": true, ".jpg": true, ".png": true,
	".tif": true, ".tiff": true, ".webp": true,
}

var imageDemuxers = map[string]bool{
	"bmp_pipe": true, "image2": true, "jpeg_pipe": true, "png_pipe": true,
	"tiff_pipe": true, "webp_pipe": true,
}

func Classify(path string, data ProbeData) Kind {
	format := primaryFormat(strings.ToLower(data.Format.FormatName))
	if format == "gif" || strings.EqualFold(filepath.Ext(path), ".gif") {
		return KindUnknown
	}

	var video, audio bool
	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "video":
			video = true
		case "audio":
			audio = true
		}
	}

	if video && !audio && (imageExtensions[strings.ToLower(filepath.Ext(path))] || imageDemuxers[format]) {
		return KindImage
	}
	if video {
		return KindVideo
	}
	if audio {
		return KindAudio
	}
	return KindUnknown
}
