package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	pdfservice "github.com/taharaLovelace/Goverter/internal/pdf"
)

func newPDFMergeCommand() *cobra.Command {
	var (
		output    string
		overwrite bool
		asJSON    bool
	)
	command := &cobra.Command{
		Use:   "merge <pdf> <pdf> [pdf...]",
		Short: "Merge PDF files",
		Long:  "Merge two or more PDF files in the order provided.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 2 {
				return usageError("pdf merge requires at least two PDF files")
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			if strings.TrimSpace(output) == "" {
				return usageError("--output is required")
			}
			if extension := strings.ToLower(filepath.Ext(output)); extension != "" && extension != ".pdf" {
				return usageError("--output must use the .pdf extension")
			}

			if !asJSON {
				fmt.Fprintln(command.ErrOrStderr(), "Merging PDFs...")
			}
			service := pdfservice.MergeService{Engine: pdfservice.PDFCPUEngine{}}
			summary, err := service.Merge(command.Context(), args, pdfservice.MergeOptions{
				Output:    output,
				Overwrite: overwrite,
			})
			if err != nil {
				return runtimeError(err)
			}

			if asJSON {
				if err := writeJSON(command.OutOrStdout(), summary); err != nil {
					return runtimeError(err)
				}
				return nil
			}
			fmt.Fprintf(command.OutOrStdout(), "Merged PDF: %s\n", summary.Output)
			fmt.Fprintf(command.OutOrStdout(), "Files: %d\n", len(summary.Files))
			fmt.Fprintf(command.OutOrStdout(), "Pages: %d\n", summary.PageCount)
			return nil
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "output PDF file (required)")
	command.Flags().BoolVar(&overwrite, "overwrite", false, "replace an existing PDF")
	command.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	return command
}
