//go:build (!cgo || pact_purego) && !pact_goffi && !windows

package native

import (
	"fmt"

	"github.com/ebitengine/purego"
)

func openLibrary(path string) (uintptr, error) {
	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, fmt.Errorf("dlopen %s: %w", path, err)
	}

	return lib, nil
}

func lookupSymbol(lib uintptr, name string) error {
	_, err := purego.Dlsym(lib, name)
	if err != nil {
		return fmt.Errorf("dlsym: %w", err)
	}

	return nil
}
