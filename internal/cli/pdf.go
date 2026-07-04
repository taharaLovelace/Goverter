package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	pdfservice "github.com/taharaLovelace/Goverter/internal/pdf"
)

func newPDFCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "pdf",
		Short: "Create and process PDF files",
	}
	command.AddCommand(newPDFImagesCommand())
	return command
}

func newPDFImagesCommand() *cobra.Command {
	var (
		output      string
		pageSize    string
		orientation string
		margin      string
		recursive   bool
		overwrite   bool
		asJSON      bool
	)
	command := &cobra.Command{
		Use:   "images <image-or-directory>...",
		Short: "Combine images into a PDF",
		Long: "Combine JPG, PNG, TIFF, or WebP images into a PDF.\n" +
			"Input arguments keep their order; images discovered in directories use natural filename order.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError("pdf images requires at least one image or directory")
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
			if _, err := pdfservice.ParseLayout(pageSize, orientation, margin); err != nil {
				return usageError("%s", err)
			}

			if !asJSON {
				fmt.Fprintln(command.ErrOrStderr(), "Creating PDF...")
			}
			service := pdfservice.Service{Engine: pdfservice.PDFCPUEngine{}}
			summary, err := service.Create(command.Context(), args, pdfservice.Options{
				Output:      output,
				PageSize:    pageSize,
				Orientation: orientation,
				Margin:      margin,
				Recursive:   recursive,
				Overwrite:   overwrite,
			})
			if err != nil {
				return runtimeError(err)
			}

			if asJSON {
				encoder := json.NewEncoder(command.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(summary); err != nil {
					return runtimeError(err)
				}
				return nil
			}
			fmt.Fprintf(command.OutOrStdout(), "Created PDF: %s\n", summary.Output)
			fmt.Fprintf(command.OutOrStdout(), "Pages: %d\n", summary.PageCount)
			return nil
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "output PDF file (required)")
	command.Flags().StringVar(&pageSize, "page-size", "a4", "page size: a4, letter, or fit")
	command.Flags().StringVar(&orientation, "orientation", "portrait", "page orientation: portrait or landscape")
	command.Flags().StringVar(&margin, "margin", "none", "page margin: none, small (10mm), or large (20mm)")
	command.Flags().BoolVarP(&recursive, "recursive", "r", false, "include images in nested directories")
	command.Flags().BoolVar(&overwrite, "overwrite", false, "replace an existing PDF")
	command.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	return command
}
