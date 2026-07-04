package pdf

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeEngine struct {
	err    error
	images []string
	layout Layout
}

func (e *fakeEngine) Create(images []string, output string, layout Layout) (int, error) {
	e.images = append([]string(nil), images...)
	e.layout = layout
	if writeErr := os.WriteFile(output, []byte("%PDF-complete"), 0o644); writeErr != nil {
		return 0, writeErr
	}
	return len(images), e.err
}

func TestServiceCreatesPDFAndPublishesOnlyCompletedOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	image := filepath.Join(directory, "photo.png")
	if err := os.WriteFile(image, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "album.pdf")
	engine := &fakeEngine{}
	service := Service{Engine: engine}

	summary, err := service.Create(context.Background(), []string{image}, Options{
		Output: output, PageSize: "a4", Orientation: "portrait", Margin: "small",
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.PageCount != 1 || summary.Output != output {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "%PDF-complete" {
		t.Fatalf("output = %q", data)
	}
	assertNoPDFTemporaryFiles(t, directory)
}

func TestServicePreservesExistingOutputWithoutOverwrite(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	image := filepath.Join(directory, "photo.jpg")
	output := filepath.Join(directory, "album.pdf")
	os.WriteFile(image, []byte("fixture"), 0o644)
	os.WriteFile(output, []byte("keep"), 0o644)

	service := Service{Engine: &fakeEngine{}}
	if _, err := service.Create(context.Background(), []string{image}, Options{
		Output: output, PageSize: "a4", Orientation: "portrait", Margin: "none",
	}); err == nil {
		t.Fatal("expected existing output to fail")
	}
	data, _ := os.ReadFile(output)
	if string(data) != "keep" {
		t.Fatalf("existing output was changed: %q", data)
	}
}

func TestServiceReplacesExistingOutputWithOverwrite(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	image := filepath.Join(directory, "photo.jpg")
	output := filepath.Join(directory, "album.pdf")
	os.WriteFile(image, []byte("fixture"), 0o644)
	os.WriteFile(output, []byte("old"), 0o644)

	service := Service{Engine: &fakeEngine{}}
	if _, err := service.Create(context.Background(), []string{image}, Options{
		Output: output, PageSize: "a4", Orientation: "portrait", Margin: "none", Overwrite: true,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "%PDF-complete" {
		t.Fatalf("output was not replaced: %q", data)
	}
	assertNoPDFTemporaryFiles(t, directory)
}

func TestServiceRemovesPartialPDFAfterEngineFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	image := filepath.Join(directory, "photo.jpg")
	output := filepath.Join(directory, "album.pdf")
	os.WriteFile(image, []byte("fixture"), 0o644)

	service := Service{Engine: &fakeEngine{err: errors.New("broken image")}}
	if _, err := service.Create(context.Background(), []string{image}, Options{
		Output: output, PageSize: "a4", Orientation: "portrait", Margin: "none",
	}); err == nil {
		t.Fatal("expected engine failure")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("final PDF exists after failure: %v", err)
	}
	assertNoPDFTemporaryFiles(t, directory)
}

func TestServiceHonorsCancellationBeforeWork(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := Service{Engine: &fakeEngine{}}
	_, err := service.Create(ctx, []string{"unused"}, Options{
		Output: "unused.pdf", PageSize: "a4", Orientation: "portrait", Margin: "none",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestOutputPathAddsPDFExtension(t *testing.T) {
	t.Parallel()
	path, err := outputPath(filepath.Join(t.TempDir(), "album"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "album.pdf") {
		t.Fatalf("output path = %q", path)
	}
}

func assertNoPDFTemporaryFiles(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".goverter-") {
			t.Errorf("temporary file was not removed: %s", entry.Name())
		}
	}
}
