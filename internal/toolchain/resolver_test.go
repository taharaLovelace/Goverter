package toolchain

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolverPrefersEnvironmentDirectory(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	ffmpeg := filepath.Join(directory, "ffmpeg"+extension)
	ffprobe := filepath.Join(directory, "ffprobe"+extension)
	os.WriteFile(ffmpeg, []byte(""), 0o644)
	os.WriteFile(ffprobe, []byte(""), 0o644)

	resolver := Resolver{
		ExecutablePath: func() (string, error) { return "", errors.New("unused") },
		LookupPath:     func(string) (string, error) { return "", errors.New("unused") },
		Getenv: func(name string) string {
			if name == envDirectory {
				return directory
			}
			return ""
		},
	}
	if got, err := resolver.FFmpeg(); err != nil || got != ffmpeg {
		t.Fatalf("FFmpeg() = %q, %v", got, err)
	}
	if got, err := resolver.FFprobe(); err != nil || got != ffprobe {
		t.Fatalf("FFprobe() = %q, %v", got, err)
	}
}

func TestResolverFallsBackToPATH(t *testing.T) {
	t.Parallel()
	resolver := Resolver{
		ExecutablePath: func() (string, error) { return "C:/app/goverter.exe", nil },
		LookupPath: func(name string) (string, error) {
			if name == "ffmpeg" {
				return "C:/ffmpeg/bin/ffmpeg.exe", nil
			}
			return "", errors.New("missing")
		},
		Getenv: func(string) string { return "" },
	}
	got, err := resolver.FFmpeg()
	if err != nil || got != "C:/ffmpeg/bin/ffmpeg.exe" {
		t.Fatalf("FFmpeg() = %q, %v", got, err)
	}
}
