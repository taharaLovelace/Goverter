package pdf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	Engine Engine
}

func (s Service) Create(ctx context.Context, inputs []string, options Options) (Summary, error) {
	layout, err := ParseLayout(options.PageSize, options.Orientation, options.Margin)
	if err != nil {
		return Summary{}, err
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	images, err := CollectImages(inputs, options.Recursive)
	if err != nil {
		return Summary{}, err
	}
	output, err := outputPath(options.Output)
	if err != nil {
		return Summary{}, err
	}
	if !options.Overwrite {
		if _, err := os.Lstat(output); err == nil {
			return Summary{}, fmt.Errorf("output already exists: %s", output)
		} else if !os.IsNotExist(err) {
			return Summary{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o750); err != nil {
		return Summary{}, fmt.Errorf("create output directory: %w", err)
	}

	temporary, err := temporaryOutputPath(output)
	if err != nil {
		return Summary{}, fmt.Errorf("create temporary output: %w", err)
	}
	defer os.Remove(temporary)

	pageCount, err := s.Engine.Create(images, temporary, layout)
	if err != nil {
		return Summary{}, err
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	if err := replaceOutput(temporary, output, options.Overwrite); err != nil {
		return Summary{}, err
	}

	return Summary{
		Output:      output,
		PageCount:   pageCount,
		Images:      images,
		PageSize:    layout.PageSize,
		Orientation: layout.Orientation,
		Margin:      layout.Margin,
	}, nil
}

func outputPath(requested string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return "", fmt.Errorf("--output is required")
	}
	output, err := filepath.Abs(requested)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	extension := strings.ToLower(filepath.Ext(output))
	switch extension {
	case "":
		output += ".pdf"
	case ".pdf":
	default:
		return "", fmt.Errorf("--output must use the .pdf extension")
	}
	return output, nil
}

func temporaryOutputPath(output string) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(output), "."+filepath.Base(output)+".goverter-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return "", errors.Join(err, os.Remove(name))
	}
	if err := os.Remove(name); err != nil {
		return "", err
	}
	return name, nil
}
