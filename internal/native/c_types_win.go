//go:build windows && cgo && !pact_purego && !pact_goffi

package native

import "C"

type CUlong = C.ulonglong
