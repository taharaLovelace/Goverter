package toolchain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const envDirectory = "GOVERTER_FFMPEG_DIR"

type Resolver struct {
	ExecutablePath func() (string, error)
	LookupPath     func(string) (string, error)
	Getenv         func(string) string
}

func NewResolver() Resolver {
	return Resolver{
		ExecutablePath: os.Executable,
		LookupPath:     exec.LookPath,
		Getenv:         os.Getenv,
	}
}

func (r Resolver) FFmpeg() (string, error) {
	return r.find("ffmpeg.exe", "ffmpeg")
}

func (r Resolver) FFprobe() (string, error) {
	return r.find("ffprobe.exe", "ffprobe")
}

func (r Resolver) find(windowsName, pathName string) (string, error) {
	if directory := r.Getenv(envDirectory); directory != "" {
		candidate := filepath.Join(directory, windowsName)
		if isFile(candidate) {
			return candidate, nil
		}
		return "", fmt.Errorf("%s is set, but %s was not found there", envDirectory, windowsName)
	}

	if executable, err := r.ExecutablePath(); err == nil {
		candidate := filepath.Join(filepath.Dir(executable), "tools", windowsName)
		if isFile(candidate) {
			return candidate, nil
		}
	}

	if found, err := r.LookupPath(pathName); err == nil {
		return found, nil
	}
	if found, err := r.LookupPath(windowsName); err == nil {
		return found, nil
	}

	return "", fmt.Errorf("%s was not found; reinstall Goverter or set %s", pathName, envDirectory)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
