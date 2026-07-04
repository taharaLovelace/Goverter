package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MergeEngine interface {
	Merge([]string, string) (int, error)
}

type MergeOptions struct {
	Output    string
	Overwrite bool
}

type MergeSummary struct {
	Output    string   `json:"output"`
	PageCount int      `json:"page_count"`
	Files     []string `json:"files"`
}

type MergeService struct {
	Engine MergeEngine
}

func (s MergeService) Merge(
	ctx context.Context,
	inputs []string,
	options MergeOptions,
) (summary MergeSummary, mergeErr error) {
	if err := ctx.Err(); err != nil {
		return MergeSummary{}, err
	}
	files, err := resolvePDFInputs(inputs)
	if err != nil {
		return MergeSummary{}, err
	}
	output, err := outputPath(options.Output)
	if err != nil {
		return MergeSummary{}, err
	}
	if !options.Overwrite {
		if _, err := os.Lstat(output); err == nil {
			return MergeSummary{}, fmt.Errorf("output already exists: %s", output)
		} else if !os.IsNotExist(err) {
			return MergeSummary{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o750); err != nil {
		return MergeSummary{}, fmt.Errorf("create output directory: %w", err)
	}

	temporary, err := temporaryOutputPath(output)
	if err != nil {
		return MergeSummary{}, fmt.Errorf("create temporary output: %w", err)
	}
	defer func() {
		if err := os.Remove(temporary); err != nil && !os.IsNotExist(err) {
			mergeErr = errors.Join(mergeErr, fmt.Errorf("remove temporary output: %w", err))
		}
	}()

	pageCount, err := s.Engine.Merge(files, temporary)
	if err != nil {
		return MergeSummary{}, err
	}
	if err := ctx.Err(); err != nil {
		return MergeSummary{}, err
	}
	if err := replaceOutput(temporary, output, options.Overwrite); err != nil {
		return MergeSummary{}, err
	}

	return MergeSummary{
		Output:    output,
		PageCount: pageCount,
		Files:     files,
	}, nil
}

func resolvePDFInputs(inputs []string) ([]string, error) {
	if len(inputs) < 2 {
		return nil, fmt.Errorf("at least two PDF files are required")
	}

	files := make([]string, 0, len(inputs))
	for _, input := range inputs {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return nil, fmt.Errorf("resolve input %q: %w", input, err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return nil, fmt.Errorf("open input %q: %w", input, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("input %q must be a regular PDF file", input)
		}
		if !strings.EqualFold(filepath.Ext(absolute), ".pdf") {
			return nil, fmt.Errorf("input %q must use the .pdf extension", input)
		}
		files = append(files, absolute)
	}
	return files, nil
}
