package integration

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfservice "github.com/taharaLovelace/Goverter/internal/pdf"
)

func TestRealImagesToPDF(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "image1.png")
	second := filepath.Join(directory, "image2.jpg")
	writePNG(t, first, 120, 80)
	writeJPEG(t, second, 80, 120)

	output := filepath.Join(directory, "album.pdf")
	service := pdfservice.Service{Engine: pdfservice.PDFCPUEngine{}}
	summary, err := service.Create(context.Background(), []string{first, second}, pdfservice.Options{
		Output:      output,
		PageSize:    "a4",
		Orientation: "landscape",
		Margin:      "small",
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.PageCount != 2 {
		t.Fatalf("page count = %d, want 2", summary.PageCount)
	}
	count, err := pdfcpuapi.PageCountFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("PDF page count = %d, want 2", count)
	}
	dimensions, err := pdfcpuapi.PageDimsFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(dimensions) != 2 {
		t.Fatalf("unexpected page dimensions: %#v", dimensions)
	}
	for page, dimension := range dimensions {
		if !dimension.Landscape() {
			t.Fatalf("page %d is not landscape: %#v", page+1, dimension)
		}
	}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, testImage(width, height)); err != nil {
		t.Fatal(err)
	}
}

func writeJPEG(t *testing.T, path string, width, height int) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, testImage(width, height), nil); err != nil {
		t.Fatal(err)
	}
}

func testImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 255),
				G: uint8(y % 255),
				B: 160,
				A: 255,
			})
		}
	}
	return img
}
