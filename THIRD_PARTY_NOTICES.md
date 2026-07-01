# Third-party notices

## FFmpeg

Goverter invokes the separate `ffmpeg` and `ffprobe` command-line programs.
The Windows installer redistributes the Gyan "release essentials" build of
FFmpeg described in `tools.lock.json`.

That build is licensed under GNU GPL version 3 and includes optional external
libraries with their own notices. The installer includes the unmodified
license and README files supplied with the pinned binary package.

- Project: https://ffmpeg.org/
- Binary provider: https://www.gyan.dev/ffmpeg/builds/
- Source revision identified by the binary provider:
  https://github.com/FFmpeg/FFmpeg/commit/38b88335f9
- Upstream FFmpeg 8.1.2 source archive:
  https://ffmpeg.org/releases/ffmpeg-8.1.2.tar.xz
- FFmpeg legal information: https://ffmpeg.org/legal.html

FFmpeg is a trademark of Fabrice Bellard, the originator of the FFmpeg
project. Goverter is not affiliated with or endorsed by the FFmpeg project or
the binary provider.

## Cobra

Goverter uses Cobra, Copyright (c) 2013-2023 Steve Francia, under the Apache
License 2.0.

- Project: https://github.com/spf13/cobra
- License: https://github.com/spf13/cobra/blob/main/LICENSE.txt

## pflag

Goverter uses pflag, Copyright (c) 2012 Alex Ogier and The Go Authors, under
the 3-Clause BSD License.

- Project: https://github.com/spf13/pflag

## mousetrap

Goverter uses mousetrap under the Apache License 2.0.

- Project: https://github.com/inconshreveable/mousetrap

The complete license texts for these dependencies are included in the
Windows installation under the `licenses` directory.

This notice is informational and is not legal advice.
