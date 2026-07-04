package pdf

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectImagesPreservesInputOrderAndNaturallySortsDirectories(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	nested := filepath.Join(directory, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"image10.jpg", "image2.png", "image1.webp", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nestedImage := filepath.Join(nested, "image3.tiff")
	if err := os.WriteFile(nestedImage, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}

	explicit := filepath.Join(directory, "image10.jpg")
	images, err := CollectImages([]string{explicit, directory}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		explicit,
		filepath.Join(directory, "image1.webp"),
		filepath.Join(directory, "image2.png"),
		filepath.Join(directory, "image10.jpg"),
	}
	if !reflect.DeepEqual(images, want) {
		t.Fatalf("images = %#v, want %#v", images, want)
	}

	recursive, err := CollectImages([]string{directory}, true)
	if err != nil {
		t.Fatal(err)
	}
	if recursive[len(recursive)-1] != nestedImage {
		t.Fatalf("nested image missing or out of order: %#v", recursive)
	}
}

func TestCollectImagesRejectsUnsupportedExplicitFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "document.txt")
	if err := os.WriteFile(path, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CollectImages([]string{path}, false); err == nil {
		t.Fatal("expected unsupported file to fail")
	}
}

func TestCollectImagesRejectsEmptyDirectory(t *testing.T) {
	t.Parallel()
	if _, err := CollectImages([]string{t.TempDir()}, false); err == nil {
		t.Fatal("expected empty directory to fail")
	}
}
