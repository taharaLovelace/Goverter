//go:build !windows

package publish

import (
	"fmt"
	"os"
)

func Replace(source, destination string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Lstat(destination); err == nil {
			return fmt.Errorf("output already exists: %s", destination)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return os.Rename(source, destination)
}
