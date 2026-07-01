package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	converter "github.com/taharaLovelace/Goverter/internal/convert"
	"github.com/taharaLovelace/Goverter/internal/media"
)

func newConvertCommand(deps dependencies) *cobra.Command {
	var (
		target     string
		output     string
		preset     string
		recursive  bool
		overwrite  bool
		asJSON     bool
		noProgress bool
	)
	command := &cobra.Command{
		Use:   "convert <input>",
		Short: "Convert a media file or directory",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return usageError("convert requires exactly one input file or directory")
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			if target == "" {
				return usageError("--to is required")
			}
			format, err := converter.LookupFormat(target)
			if err != nil {
				return usageError("%s", err)
			}
			if _, err := converter.ParsePreset(preset); err != nil {
				return usageError("%s", err)
			}
			if format.Lossless && command.Flags().Changed("preset") {
				return usageError("--preset does not apply to lossless %s output", format.Name)
			}

			ffprobePath, err := deps.resolver.FFprobe()
			if err != nil {
				return runtimeError(err)
			}
			ffmpegPath, err := deps.resolver.FFmpeg()
			if err != nil {
				return runtimeError(err)
			}

			service := converter.Service{
				Prober: media.FFprobe{Path: ffprobePath},
				Runner: converter.FFmpegRunner{Path: ffmpegPath},
			}
			var progress converter.ProgressFunc
			var reporter *progressReporter
			if !noProgress {
				reporter = newProgressReporter(command.ErrOrStderr())
				progress = reporter.Update
			}
			summary, convertErr := service.Convert(command.Context(), args[0], converter.Options{
				Target:         target,
				Output:         output,
				Preset:         preset,
				PresetExplicit: command.Flags().Changed("preset"),
				Recursive:      recursive,
				Overwrite:      overwrite,
			}, progress)
			if reporter != nil {
				reporter.Finish()
			}

			if asJSON {
				encoder := json.NewEncoder(command.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(summary); err != nil {
					return runtimeError(err)
				}
			} else {
				printSummary(command.OutOrStdout(), summary)
			}
			if convertErr != nil {
				if errors.Is(convertErr, os.ErrInvalid) {
					return usageError("%s", convertErr)
				}
				return runtimeError(convertErr)
			}
			return nil
		},
	}
	command.Flags().StringVar(&target, "to", "", "output format (required)")
	command.Flags().StringVarP(&output, "output", "o", "", "output file or directory")
	command.Flags().StringVar(&preset, "preset", string(converter.PresetBalanced), "quality preset")
	command.Flags().BoolVarP(&recursive, "recursive", "r", false, "include nested directories")
	command.Flags().BoolVar(&overwrite, "overwrite", false, "replace existing output files")
	command.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	command.Flags().BoolVar(&noProgress, "no-progress", false, "disable conversion progress")
	return command
}

func printSummary(writer io.Writer, summary converter.Summary) {
	for _, result := range summary.Results {
		switch result.Status {
		case converter.StatusSucceeded:
			fmt.Fprintf(writer, "Converted: %s -> %s\n", result.Input, result.Output)
		case converter.StatusFailed:
			fmt.Fprintf(writer, "Failed:    %s (%s)\n", result.Input, result.Error)
		case converter.StatusSkipped:
			fmt.Fprintf(writer, "Skipped:   %s (%s)\n", result.Input, result.Error)
		}
	}
	if summary.Total > 0 {
		fmt.Fprintf(
			writer,
			"Summary: %d succeeded, %d failed, %d skipped\n",
			summary.Succeeded,
			summary.Failed,
			summary.Skipped,
		)
	}
}

type progressReporter struct {
	writer      io.Writer
	interactive bool
	active      bool
	current     int
	lastBucket  int
}

func newProgressReporter(writer io.Writer) *progressReporter {
	reporter := &progressReporter{writer: writer, lastBucket: -1}
	if file, ok := writer.(*os.File); ok {
		if info, err := file.Stat(); err == nil {
			reporter.interactive = info.Mode()&os.ModeCharDevice != 0
		}
	}
	return reporter
}

func (r *progressReporter) Update(progress converter.Progress) {
	name := filepath.Base(progress.Input)
	if r.active && r.current != progress.Current {
		if r.interactive {
			fmt.Fprintln(r.writer)
		}
		r.active = false
	}
	if !r.active {
		r.active = true
		r.current = progress.Current
		r.lastBucket = -1
		if !r.interactive {
			fmt.Fprintf(r.writer, "Converting [%d/%d] %s\n", progress.Current, progress.Total, name)
			return
		}
	}
	if !r.interactive {
		return
	}
	bucket := int(progress.Percent) / 2
	if bucket == r.lastBucket && !progress.Done {
		return
	}
	r.lastBucket = bucket
	fmt.Fprintf(
		r.writer,
		"\rConverting [%d/%d] %-30s %6.1f%%",
		progress.Current,
		progress.Total,
		truncate(name, 30),
		progress.Percent,
	)
	if progress.Done {
		fmt.Fprintln(r.writer)
		r.active = false
	}
}

func (r *progressReporter) Finish() {
	if r.interactive && r.active {
		fmt.Fprintln(r.writer)
		r.active = false
	}
}

func truncate(value string, length int) string {
	runes := []rune(value)
	if len(runes) <= length {
		return value
	}
	if length <= 1 {
		return string(runes[:length])
	}
	return string(runes[:length-1]) + "…"
}
