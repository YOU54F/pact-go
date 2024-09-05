//go:build (!cgo || pact_purego) && !pact_goffi && windows

package native

import "golang.org/x/sys/windows"

func openLibrary(path string) (uintptr, error) {
	handle, err := windows.LoadLibrary(path)

	return uintptr(handle), err
}
