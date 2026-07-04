package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

var supportedImageExtensions = map[string]bool{
	".jpeg": true,
	".jpg":  true,
	".png":  true,
	".tif":  true,
	".tiff": true,
	".webp": true,
}

func CollectImages(inputs []string, recursive bool) ([]string, error) {
	var images []string
	for _, input := range inputs {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return nil, fmt.Errorf("resolve input %q: %w", input, err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return nil, fmt.Errorf("open input %q: %w", input, err)
		}
		if info.Mode().IsRegular() {
			if !isSupportedImage(absolute) {
				return nil, fmt.Errorf("unsupported image %q; use JPG, PNG, TIFF, or WebP", input)
			}
			images = append(images, absolute)
			continue
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("input %q must be a regular file or directory", input)
		}

		found, err := collectDirectory(absolute, recursive)
		if err != nil {
			return nil, err
		}
		images = append(images, found...)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no supported images found")
	}
	return images, nil
}

func collectDirectory(root string, recursive bool) ([]string, error) {
	var images []string
	if !recursive {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, fmt.Errorf("read directory %q: %w", root, err)
		}
		for _, entry := range entries {
			if entry.Type().IsRegular() {
				path := filepath.Join(root, entry.Name())
				if isSupportedImage(path) {
					images = append(images, path)
				}
			}
		}
	} else {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type().IsRegular() && isSupportedImage(path) {
				images = append(images, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk directory %q: %w", root, err)
		}
	}

	sort.SliceStable(images, func(i, j int) bool {
		left, _ := filepath.Rel(root, images[i])
		right, _ := filepath.Rel(root, images[j])
		return naturalLess(left, right)
	})
	return images, nil
}

func isSupportedImage(path string) bool {
	return supportedImageExtensions[strings.ToLower(filepath.Ext(path))]
}

func naturalLess(left, right string) bool {
	l := []rune(strings.ToLower(left))
	r := []rune(strings.ToLower(right))
	for li, ri := 0, 0; li < len(l) && ri < len(r); {
		if unicode.IsDigit(l[li]) && unicode.IsDigit(r[ri]) {
			ln, nextL := numberToken(l, li)
			rn, nextR := numberToken(r, ri)
			if len(ln) != len(rn) {
				return len(ln) < len(rn)
			}
			if string(ln) != string(rn) {
				return string(ln) < string(rn)
			}
			li, ri = nextL, nextR
			continue
		}
		if l[li] != r[ri] {
			return l[li] < r[ri]
		}
		li++
		ri++
	}
	return len(l) < len(r)
}

func numberToken(value []rune, start int) ([]rune, int) {
	end := start
	for end < len(value) && unicode.IsDigit(value[end]) {
		end++
	}
	token := value[start:end]
	for len(token) > 1 && token[0] == '0' {
		token = token[1:]
	}
	return token, end
}
