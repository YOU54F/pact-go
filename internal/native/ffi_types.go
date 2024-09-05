package native

import "unsafe"

// The pact_ffi library can be bound in three ways, selected at build time:
//
//   - cgo (ffi_cgo.go): the default when cgo is enabled. Links libpact_ffi
//     at build time, exactly as pact-go has always done.
//   - purego (ffi_purego.go): the default when cgo is disabled
//     (CGO_ENABLED=0), or forced with -tags pact_purego. Loads libpact_ffi
//     at runtime with github.com/ebitengine/purego; no C toolchain needed.
//   - goffi (ffi_goffi.go): opt-in with -tags pact_goffi. Loads libpact_ffi
//     at runtime with github.com/go-webgpu/goffi; no C toolchain needed.
//
// Every backend provides the same set of pactffi_* functions with the same
// Go signatures, so the rest of this package does not know which is in use.
// The mapping from the C prototypes in pact.h is:
//
//	const char * (in)        string (copied, NUL terminated for the call)
//	const char * (nullable)  *byte, see cstr
//	char * (returned)        unsafe.Pointer, see goString and libRustFree
//	opaque pointers          unsafe.Pointer
//	PactHandle               uint16
//	InteractionHandle        uint32
//	int, int32_t             int32
//	unsigned short           uint16
//	size_t                   uintptr
//	unsigned char            uint8
//	bool                     bool

// goString copies the NUL-terminated C string at p into a Go string. It
// returns "" if p is nil. The C string is not freed.
func goString(p unsafe.Pointer) string {
	if p == nil {
		return ""
	}

	n := 0
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}

	return string(unsafe.Slice((*byte)(p), n))
}

// goBytes copies n bytes of C memory starting at p into a Go slice.
func goBytes(p unsafe.Pointer, n int) []byte {
	if p == nil || n <= 0 {
		return []byte{}
	}

	out := make([]byte, n)
	copy(out, unsafe.Slice((*byte)(p), n))

	return out
}

// cstr returns a pointer to a NUL-terminated copy of s, for the C string
// arguments that may also be passed as NULL (a nil *byte).
func cstr(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)

	return &b[0]
}
