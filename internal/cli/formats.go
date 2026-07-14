package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	converter "github.com/taharaLovelace/Goverter/internal/convert"
	"github.com/taharaLovelace/Goverter/internal/media"
)

type formatsOutput struct {
	Presets []converter.Preset `json:"presets"`
	Formats []converter.Format `json:"formats"`
}

func newFormatsCommand() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{
		Use:   "formats",
		Short: "List supported output formats",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			output := formatsOutput{
				Presets: converter.Presets,
				Formats: converter.ListFormats(),
			}
			if asJSON {
				return writeJSON(command.OutOrStdout(), output)
			}

			out := command.OutOrStdout()
			for _, kind := range []media.Kind{media.KindVideo, media.KindAudio, media.KindImage} {
				fmt.Fprintf(out, "%s:\n", kind)
				for _, format := range output.Formats {
					if format.Kind != kind {
						continue
					}
					lossless := ""
					if format.Lossless {
						lossless = " [lossless]"
					}
					fmt.Fprintf(out, "  %-5s %s%s\n", format.Name, format.Codecs, lossless)
				}
			}
			fmt.Fprintln(out, "\nPresets: compact, balanced (default), quality")
			fmt.Fprintln(out, "Lossless formats do not accept an explicit preset.")
			return nil
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	return command
}
