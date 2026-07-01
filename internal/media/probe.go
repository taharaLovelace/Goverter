package media

import (
	"context"
	"fmt"
	"os/exec"
)

type Prober interface {
	Probe(context.Context, string) (Info, error)
}

type FFprobe struct {
	Path string
}

func (p FFprobe) Probe(ctx context.Context, path string) (Info, error) {
	command := exec.CommandContext(
		ctx,
		p.Path,
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-of", "json",
		path,
	)
	output, err := command.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return Info{}, fmt.Errorf("ffprobe failed: %s", cleanMessage(exitError.Stderr))
		}
		return Info{}, fmt.Errorf("run ffprobe: %w", err)
	}
	return ParseProbeJSON(output, path)
}

func cleanMessage(data []byte) string {
	if len(data) == 0 {
		return "input is not recognized as media"
	}
	return string(data)
}
