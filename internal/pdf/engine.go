package pdf

import (
	"fmt"
	"sync"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Engine interface {
	Create([]string, string, Layout) (int, error)
}

type PDFCPUEngine struct{}

var disablePDFCPUConfig sync.Once

func (PDFCPUEngine) Create(images []string, output string, layout Layout) (int, error) {
	config := pdfcpuConfiguration()

	configuration, err := layout.ImportConfiguration()
	if err != nil {
		return 0, fmt.Errorf("configure PDF layout: %w", err)
	}
	if err := pdfcpuapi.ImportImagesFile(images, output, configuration, config); err != nil {
		return 0, fmt.Errorf("create PDF: %w", err)
	}
	if err := pdfcpuapi.ValidateFile(output, config); err != nil {
		return 0, fmt.Errorf("validate generated PDF: %w", err)
	}
	return pageCount(output)
}

func (PDFCPUEngine) Merge(files []string, output string) (int, error) {
	config := pdfcpuConfiguration()
	config.CreateBookmarks = false

	if err := pdfcpuapi.MergeCreateFile(files, output, false, config); err != nil {
		return 0, fmt.Errorf("merge PDFs: %w", err)
	}
	if err := pdfcpuapi.ValidateFile(output, config); err != nil {
		return 0, fmt.Errorf("validate merged PDF: %w", err)
	}
	return pageCount(output)
}

func pdfcpuConfiguration() *model.Configuration {
	disablePDFCPUConfig.Do(pdfcpuapi.DisableConfigDir)
	return model.NewDefaultConfiguration()
}

func pageCount(output string) (int, error) {
	pageCount, err := pdfcpuapi.PageCountFile(output)
	if err != nil {
		return 0, fmt.Errorf("count generated PDF pages: %w", err)
	}
	return pageCount, nil
}
