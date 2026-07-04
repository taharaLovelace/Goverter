package pdf

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeMergeEngine struct {
	err   error
	files []string
	pages int
}

func (e *fakeMergeEngine) Merge(files []string, output string) (int, error) {
	e.files = append([]string(nil), files...)
	if err := os.WriteFile(output, []byte("%PDF-merged"), 0o644); err != nil {
		return 0, err
	}
	return e.pages, e.err
}

func TestResolvePDFInputsPreservesExplicitOrder(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := writePDFMergeFixture(t, directory, "first.PDF")
	second := writePDFMergeFixture(t, directory, "second.pdf")

	files, err := resolvePDFInputs([]string{second, first})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0] != second || files[1] != first {
		t.Fatalf("unexpected input order: %#v", files)
	}
}

func TestResolvePDFInputsRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	pdfFile := writePDFMergeFixture(t, directory, "document.pdf")
	textFile := writePDFMergeFixture(t, directory, "document.txt")

	tests := []struct {
		name   string
		inputs []string
	}{
		{name: "fewer than two", inputs: []string{pdfFile}},
		{name: "wrong extension", inputs: []string{pdfFile, textFile}},
		{name: "missing file", inputs: []string{pdfFile, filepath.Join(directory, "missing.pdf")}},
		{name: "directory", inputs: []string{pdfFile, directory}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := resolvePDFInputs(test.inputs); err == nil {
				t.Fatal("expected invalid input to fail")
			}
		})
	}
}

func TestMergeServicePublishesCompletedOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := writePDFMergeFixture(t, directory, "first.pdf")
	second := writePDFMergeFixture(t, directory, "second.pdf")
	output := filepath.Join(directory, "nested", "combined")
	engine := &fakeMergeEngine{pages: 3}

	summary, err := (MergeService{Engine: engine}).Merge(
		context.Background(),
		[]string{second, first},
		MergeOptions{Output: output},
	)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := output + ".pdf"
	if summary.Output != expectedOutput || summary.PageCount != 3 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if len(summary.Files) != 2 || summary.Files[0] != second || summary.Files[1] != first {
		t.Fatalf("unexpected files: %#v", summary.Files)
	}
	data, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "%PDF-merged" {
		t.Fatalf("output = %q", data)
	}
	assertNoPDFTemporaryFiles(t, filepath.Dir(expectedOutput))
}

func TestMergeServiceProtectsAndOverwritesExistingOutput(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := writePDFMergeFixture(t, directory, "first.pdf")
	second := writePDFMergeFixture(t, directory, "second.pdf")
	output := filepath.Join(directory, "combined.pdf")
	if err := os.WriteFile(output, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	service := MergeService{Engine: &fakeMergeEngine{pages: 2}}

	if _, err := service.Merge(context.Background(), []string{first, second}, MergeOptions{
		Output: output,
	}); err == nil {
		t.Fatal("expected existing output to fail")
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep" {
		t.Fatalf("existing output was changed: %q", data)
	}

	if _, err := service.Merge(context.Background(), []string{first, second}, MergeOptions{
		Output: output, Overwrite: true,
	}); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "%PDF-merged" {
		t.Fatalf("output was not replaced: %q", data)
	}
	assertNoPDFTemporaryFiles(t, directory)
}

func TestMergeServiceRemovesPartialOutputAfterFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := writePDFMergeFixture(t, directory, "first.pdf")
	second := writePDFMergeFixture(t, directory, "second.pdf")
	output := filepath.Join(directory, "combined.pdf")
	service := MergeService{Engine: &fakeMergeEngine{
		err: errors.New("corrupt PDF"), pages: 2,
	}}

	if _, err := service.Merge(context.Background(), []string{first, second}, MergeOptions{
		Output: output,
	}); err == nil {
		t.Fatal("expected merge failure")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("final PDF exists after failure: %v", err)
	}
	assertNoPDFTemporaryFiles(t, directory)
}

func TestMergeServiceHonorsCancellationBeforeWork(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (MergeService{Engine: &fakeMergeEngine{}}).Merge(
		ctx,
		[]string{"unused", "unused"},
		MergeOptions{Output: "unused.pdf"},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestMergeSummaryJSONFieldNames(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(MergeSummary{
		Output: "combined.pdf", PageCount: 2, Files: []string{"one.pdf", "two.pdf"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["output"] != "combined.pdf" || payload["page_count"] != float64(2) {
		t.Fatalf("unexpected JSON: %s", data)
	}
	if _, ok := payload["files"]; !ok {
		t.Fatalf("files missing from JSON: %s", data)
	}
}

func writePDFMergeFixture(t *testing.T, directory, name string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte("%PDF-fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
