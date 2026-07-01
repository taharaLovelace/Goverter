package convert

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type Runner interface {
	Run(context.Context, Plan, func(float64, bool)) error
}

type FFmpegRunner struct {
	Path string
}

func (r FFmpegRunner) Run(ctx context.Context, plan Plan, progress func(float64, bool)) error {
	command := exec.CommandContext(ctx, r.Path, plan.Args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("read ffmpeg progress: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return fmt.Errorf("read ffmpeg errors: %w", err)
	}

	if err := command.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	messages := &messageBuffer{limit: 20}
	var stderrErr error
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			messages.Add(scanner.Text())
		}
		stderrErr = scanner.Err()
	}()

	progressErr := ParseProgress(stdout, plan.Duration, progress)
	commandErr := command.Wait()
	wait.Wait()

	if errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	if progressErr != nil {
		return fmt.Errorf("read ffmpeg progress: %w", progressErr)
	}
	if stderrErr != nil {
		return fmt.Errorf("read ffmpeg errors: %w", stderrErr)
	}
	if commandErr != nil {
		message := messages.String()
		if message == "" {
			message = commandErr.Error()
		}
		return fmt.Errorf("ffmpeg failed:\n%s", message)
	}
	return nil
}

type messageBuffer struct {
	mu    sync.Mutex
	lines []string
	limit int
}

func (b *messageBuffer) Add(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = append(b.lines, line)
	if len(b.lines) > b.limit {
		b.lines = b.lines[len(b.lines)-b.limit:]
	}
}

func (b *messageBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.Join(b.lines, "\n")
}
