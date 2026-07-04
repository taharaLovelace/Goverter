# Goverter

Goverter is a small, script-friendly CLI for converting video, audio, and
image files with FFmpeg and combining images into PDF documents. It offers
safe defaults and focused commands instead of exposing every underlying
option.

## Features

- Convert one file or a complete directory.
- Recursively preserve a directory tree with `--recursive`.
- Extract audio from video.
- Inspect media metadata with `ffprobe`.
- Combine JPG, PNG, TIFF, and WebP images into one PDF.
- Human-readable progress and stable JSON results.
- Never overwrite an existing output unless `--overwrite` is supplied.
- Write to a temporary file and publish the final output only after FFmpeg
  succeeds.

## Supported output formats

| Media | Formats |
| --- | --- |
| Video | MP4 (H.264/AAC), WebM (VP9/Opus) |
| Audio | MP3, WAV, FLAC, OGG/Vorbis |
| Image | JPG, PNG, WebP |

Goverter supports conversions within one media category and video-to-audio
extraction. GIF, AVIF, multiple stream selection, hardware acceleration, and
custom presets are not part of the first release.

## Installation

Download the Windows x64 installer from the GitHub Releases page. The
installer includes `ffmpeg.exe` and `ffprobe.exe` and adds Goverter to the
current user's `PATH`.

For development, install FFmpeg separately or point Goverter to a directory
containing both executables:

```powershell
$env:GOVERTER_FFMPEG_DIR = "C:\path\to\ffmpeg\bin"
```

Goverter searches `GOVERTER_FFMPEG_DIR` first, then a `tools` directory next
to its own executable, and finally `PATH`.

## Usage

```text
goverter convert video.mov --to mp4
goverter convert video.mp4 --to mp3 --preset quality
goverter convert .\media --to webm --recursive
goverter convert photo.png --to webp --output photo-small.webp
goverter info video.mp4
goverter info video.mp4 --json
goverter formats
goverter pdf images cover.jpg page2.png --output album.pdf
goverter pdf images .\scans --output scans.pdf --page-size a4 --margin small
goverter pdf images .\photos --output photos.pdf --page-size fit --recursive
goverter completion powershell
```

Use `--overwrite` to replace existing results. For a single input, the
default output is placed next to that file. Directory conversions are written
under `<input>\converted`.

The `pdf images` command keeps explicit inputs in the order provided and
sorts images found in each directory by natural filename order (`page2`
before `page10`). Page sizes may be `a4`, `letter`, or `fit`; the latter makes
each PDF page match its source image. Fixed page sizes also support portrait
or landscape orientation and margins of `none`, `small` (10 mm), or `large`
(20 mm). Existing PDFs are protected unless `--overwrite` is supplied.

Progress and diagnostics are written to `stderr`; results are written to
`stdout`. `convert`, `info`, `formats`, and `pdf images` support
machine-readable JSON.

Exit codes are:

- `0`: success
- `1`: conversion, PDF generation, probing, or file-system failure
- `2`: invalid command usage
- `130`: canceled by the user

Run `goverter <command> --help` for every option.

## Development

Go 1.26.4 or newer is required.

```powershell
go mod download
go test ./...
go vet ./...
go build -o dist\goverter.exe .\cmd\goverter
```

To stage the pinned FFmpeg build used by the installer:

```powershell
.\scripts\prepare-ffmpeg.ps1
```

To run real FFmpeg integration tests after staging:

```powershell
$env:GOVERTER_FFMPEG_DIR = "$PWD\dist\tools"
go test -tags integration ./integration
```

## Licenses

Goverter is licensed under the MIT License. Release installers redistribute
an independent GPLv3 FFmpeg build. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)
and the license files included by the installer. Release history is available
in [CHANGELOG.md](CHANGELOG.md).
