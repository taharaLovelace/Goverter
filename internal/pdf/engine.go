package pdf

import (
	"fmt"
	"os"
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
	disablePDFCPUConfig.Do(pdfcpuapi.DisableConfigDir)
	config := model.NewDefaultConfiguration()

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
	file, err := os.Open(output)
	if err != nil {
		return 0, fmt.Errorf("open generated PDF: %w", err)
	}
	defer file.Close()
	pageCount, err := pdfcpuapi.PageCount(file, config)
	if err != nil {
		return 0, fmt.Errorf("count generated PDF pages: %w", err)
	}
	return pageCount, nil
}
