//go:build (!cgo || pact_purego) && !pact_goffi

package native

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// ffiBackend names the binding in use, see ffi_types.go.
const ffiBackend = "purego"

// bindLibrary loads the pact_ffi library at path and binds every symbol
// to its Go func variable.
func bindLibrary(path string, symbols []ffiSymbol) error {
	lib, err := openLibrary(path)
	if err != nil {
		return err
	}

	// RegisterLibFunc panics on a missing symbol, so check them all first.
	for _, s := range symbols {
		_, err := purego.Dlsym(lib, s.name)
		if err != nil {
			return fmt.Errorf("symbol %s: %w", s.name, err)
		}
	}
	for _, s := range symbols {
		purego.RegisterLibFunc(s.fptr, lib, s.name)
	}

	return nil
}
