package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/taharaLovelace/Goverter/internal/media"
	"github.com/taharaLovelace/Goverter/internal/toolchain"
)

func newInfoCommand(resolver toolchain.Resolver) *cobra.Command {
	var asJSON bool
	command := &cobra.Command{
		Use:   "info <file>",
		Short: "Inspect a media file",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return usageError("info requires exactly one file")
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			path, err := filepath.Abs(args[0])
			if err != nil {
				return runtimeError(err)
			}
			stat, err := os.Stat(path)
			if err != nil {
				return runtimeError(fmt.Errorf("open input: %w", err))
			}
			if !stat.Mode().IsRegular() {
				return usageError("info input must be a regular file")
			}
			probePath, err := resolver.FFprobe()
			if err != nil {
				return runtimeError(err)
			}
			info, err := (media.FFprobe{Path: probePath}).Probe(command.Context(), path)
			if err != nil {
				return runtimeError(err)
			}
			if asJSON {
				if err := writeJSON(command.OutOrStdout(), info); err != nil {
					return runtimeError(err)
				}
				return nil
			}
			printInfo(command, info)
			return nil
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	return command
}

func printInfo(command *cobra.Command, info media.Info) {
	out := command.OutOrStdout()
	fmt.Fprintf(out, "Path:     %s\n", info.Path)
	fmt.Fprintf(out, "Type:     %s\n", info.Kind)
	fmt.Fprintf(out, "Format:   %s\n", info.Format)
	if info.DurationSeconds > 0 {
		fmt.Fprintf(out, "Duration: %.3f s\n", info.DurationSeconds)
	}
	if info.SizeBytes > 0 {
		fmt.Fprintf(out, "Size:     %d bytes\n", info.SizeBytes)
	}
	fmt.Fprintln(out, "Streams:")
	for _, stream := range info.Streams {
		description := fmt.Sprintf("  #%d %s: %s", stream.Index, stream.Type, stream.Codec)
		if stream.Width > 0 && stream.Height > 0 {
			description += fmt.Sprintf(" %dx%d", stream.Width, stream.Height)
		}
		if stream.SampleRate > 0 {
			description += fmt.Sprintf(" %d Hz", stream.SampleRate)
		}
		if stream.Channels > 0 {
			description += fmt.Sprintf(" %d channel(s)", stream.Channels)
		}
		fmt.Fprintln(out, description)
	}
}
