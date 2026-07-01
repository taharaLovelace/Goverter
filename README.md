# Goverter

Goverter is a small, script-friendly CLI for converting video, audio, and
image files with FFmpeg. It offers safe defaults and a deliberately limited
set of output formats instead of exposing the entire FFmpeg command line.

## Features

- Convert one file or a complete directory.
- Recursively preserve a directory tree with `--recursive`.
- Extract audio from video.
- Inspect media metadata with `ffprobe`.
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
goverter completion powershell
```

Use `--overwrite` to replace existing results. For a single input, the
default output is placed next to that file. Directory conversions are written
under `<input>\converted`.

Progress and diagnostics are written to `stderr`; results are written to
`stdout`. `convert`, `info`, and `formats` support machine-readable JSON.

Exit codes are:

- `0`: success
- `1`: conversion, probing, or file-system failure
- `2`: invalid command usage
- `130`: canceled by the user

Run `goverter <command> --help` for every option.

## Development

Go 1.26 or newer is required.

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
