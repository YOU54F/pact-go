//go:build (!cgo || pact_purego) && !pact_goffi && windows

package native

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func openLibrary(path string) (uintptr, error) {
	handle, err := windows.LoadLibrary(path)
	if err != nil {
		return 0, fmt.Errorf("load %s: %w", path, err)
	}

	return uintptr(handle), nil
}

func lookupSymbol(lib uintptr, name string) error {
	_, err := windows.GetProcAddress(windows.Handle(lib), name)
	if err != nil {
		return fmt.Errorf("lookup %s: %w", name, err)
	}

	return nil
}
