//go:build windows

package convert

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func replaceFile(source, destination string, overwrite bool) error {
	sourcePointer, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationPointer, err := windows.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	flags := uint32(0)
	if overwrite {
		flags |= windows.MOVEFILE_REPLACE_EXISTING
	}
	if err := windows.MoveFileEx(sourcePointer, destinationPointer, flags); err != nil {
		return fmt.Errorf("move completed output: %w", err)
	}
	return nil
}
