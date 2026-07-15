package convert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/taharaLovelace/Goverter/internal/media"
	"github.com/taharaLovelace/Goverter/internal/publish"
)

type Options struct {
	Target         string
	Output         string
	Preset         string
	PresetExplicit bool
	Recursive      bool
	Overwrite      bool
}

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusSkipped   Status = "skipped"
)

type Result struct {
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Summary struct {
	Input        string   `json:"input"`
	OutputFormat string   `json:"output_format"`
	Preset       Preset   `json:"preset"`
	Total        int      `json:"total"`
	Succeeded    int      `json:"succeeded"`
	Failed       int      `json:"failed"`
	Skipped      int      `json:"skipped"`
	Results      []Result `json:"results"`
}

type Service struct {
	Prober media.Prober
	Runner Runner
}

func (s Service) Convert(ctx context.Context, input string, options Options, progress ProgressFunc) (Summary, error) {
	format, err := LookupFormat(options.Target)
	if err != nil {
		return Summary{}, err
	}
	preset, err := ParsePreset(options.Preset)
	if err != nil {
		return Summary{}, err
	}
	if format.Lossless && options.PresetExplicit {
		return Summary{}, fmt.Errorf("--preset does not apply to lossless %s output", format.Name)
	}

	absoluteInput, err := filepath.Abs(input)
	if err != nil {
		return Summary{}, fmt.Errorf("resolve input path: %w", err)
	}
	info, err := os.Stat(absoluteInput)
	if err != nil {
		return Summary{}, fmt.Errorf("open input: %w", err)
	}

	summary := Summary{
		Input:        absoluteInput,
		OutputFormat: format.Name,
		Preset:       preset,
		Results:      []Result{},
	}
	if info.IsDir() {
		err = s.convertDirectory(ctx, absoluteInput, format, preset, options, &summary, progress)
	} else if info.Mode().IsRegular() {
		err = s.convertFile(ctx, absoluteInput, format, preset, options, &summary, progress)
	} else {
		err = fmt.Errorf("input must be a regular file or directory")
	}
	return summary, err
}

func (s Service) convertFile(
	ctx context.Context,
	input string,
	format Format,
	preset Preset,
	options Options,
	summary *Summary,
	progress ProgressFunc,
) error {
	output, err := singleOutputPath(input, options.Output, format)
	if err != nil {
		return err
	}
	item, probeErr := s.Prober.Probe(ctx, input)
	if probeErr != nil {
		return fmt.Errorf("inspect %s: %w", input, probeErr)
	}
	if item.Kind == media.KindUnknown {
		return fmt.Errorf("input is not a supported video, audio, or image")
	}
	if !Compatible(item.Kind, format.Kind) {
		return fmt.Errorf("cannot convert %s input to %s output", item.Kind, format.Kind)
	}

	summary.Total = 1
	result := s.runOne(ctx, item, output, format, preset, options.Overwrite, 1, 1, progress)
	summary.Results = append(summary.Results, result)
	summary.add(result)
	if result.Status == StatusFailed {
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return errors.New(result.Error)
	}
	return nil
}

func (s Service) convertDirectory(
	ctx context.Context,
	input string,
	format Format,
	preset Preset,
	options Options,
	summary *Summary,
	progress ProgressFunc,
) error {
	outputRoot := options.Output
	if outputRoot == "" {
		outputRoot = filepath.Join(input, "converted")
	} else if !filepath.IsAbs(outputRoot) {
		outputRoot, _ = filepath.Abs(outputRoot)
	}

	files, err := collectFiles(input, outputRoot, options.Recursive)
	if err != nil {
		return err
	}
	summary.Total = len(files)
	type work struct {
		info   media.Info
		output string
	}
	var compatible []work

	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		item, probeErr := s.Prober.Probe(ctx, path)
		if probeErr != nil || item.Kind == media.KindUnknown || !Compatible(item.Kind, format.Kind) {
			result := Result{Input: path, Status: StatusSkipped}
			if probeErr == nil && item.Kind != media.KindUnknown {
				result.Error = fmt.Sprintf("incompatible %s input", item.Kind)
			} else {
				result.Error = "not recognized as supported media"
			}
			summary.Results = append(summary.Results, result)
			summary.add(result)
			continue
		}
		relative, relErr := filepath.Rel(input, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(outputRoot, strings.TrimSuffix(relative, filepath.Ext(relative))+"."+format.Name)
		compatible = append(compatible, work{info: item, output: target})
	}

	if len(compatible) == 0 {
		return fmt.Errorf("no compatible media files found")
	}

	for index, item := range compatible {
		if err := ctx.Err(); err != nil {
			return err
		}
		result := s.runOne(ctx, item.info, item.output, format, preset, options.Overwrite, index+1, len(compatible), progress)
		summary.Results = append(summary.Results, result)
		summary.add(result)
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
	}
	if summary.Failed > 0 {
		return fmt.Errorf("%d conversion(s) failed", summary.Failed)
	}
	return nil
}

func (s Service) runOne(
	ctx context.Context,
	input media.Info,
	output string,
	format Format,
	preset Preset,
	overwrite bool,
	current, total int,
	progress ProgressFunc,
) Result {
	result := Result{Input: input.Path, Output: output, Status: StatusFailed}

	if !overwrite {
		if _, err := os.Lstat(output); err == nil {
			result.Error = "output already exists"
			return result
		} else if !os.IsNotExist(err) {
			result.Error = err.Error()
			return result
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o750); err != nil {
		result.Error = fmt.Sprintf("create output directory: %v", err)
		return result
	}

	temporary, err := temporaryPath(output)
	if err != nil {
		result.Error = fmt.Sprintf("create temporary output: %v", err)
		return result
	}
	defer os.Remove(temporary)

	plan, err := BuildPlan(input, temporary, format, preset)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if progress != nil {
		progress(Progress{Input: input.Path, Current: current, Total: total})
	}
	err = s.Runner.Run(ctx, plan, func(percent float64, done bool) {
		if progress != nil {
			progress(Progress{
				Input: input.Path, Current: current, Total: total,
				Percent: percent, Done: done,
			})
		}
	})
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if err := publish.Replace(temporary, output, overwrite); err != nil {
		result.Error = err.Error()
		return result
	}
	result.Status = StatusSucceeded
	return result
}

func (s *Summary) add(result Result) {
	switch result.Status {
	case StatusSucceeded:
		s.Succeeded++
	case StatusFailed:
		s.Failed++
	case StatusSkipped:
		s.Skipped++
	}
}

func singleOutputPath(input, requested string, format Format) (string, error) {
	if requested == "" {
		return strings.TrimSuffix(input, filepath.Ext(input)) + "." + format.Name, nil
	}
	output, err := filepath.Abs(requested)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	if info, statErr := os.Stat(output); statErr == nil && info.IsDir() {
		return "", fmt.Errorf("--output must be a file path when input is a file")
	}
	extension := strings.ToLower(filepath.Ext(output))
	if extension == "" {
		output += "." + format.Name
	} else if extension != "."+format.Name && !(format.Name == "jpg" && extension == ".jpeg") {
		return "", fmt.Errorf("--output extension must match --to %s", format.Name)
	}
	return output, nil
}

func collectFiles(root, outputRoot string, recursive bool) ([]string, error) {
	absoluteOutput, _ := filepath.Abs(outputRoot)
	var files []string

	if !recursive {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			files = append(files, filepath.Join(root, entry.Name()))
		}
	} else {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				absolutePath, _ := filepath.Abs(path)
				if path != root && samePath(absolutePath, absoluteOutput) {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type().IsRegular() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

func samePath(first, second string) bool {
	first, second = filepath.Clean(first), filepath.Clean(second)
	return first == second || runtime.GOOS == "windows" && strings.EqualFold(first, second)
}

func temporaryPath(output string) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(output), "."+filepath.Base(output)+".goverter-*.tmp")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if closeErr := file.Close(); closeErr != nil {
		return "", errors.Join(closeErr, os.Remove(name))
	}
	return name, nil
}
