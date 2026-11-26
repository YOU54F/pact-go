//go:build (!cgo || pact_purego || pact_goffi) && linux

package native

import (
	"fmt"
	"unsafe"
)

// verifierStackSize is the stack size of the thread a verification runs on.
const verifierStackSize = 8 << 20

// The pthread functions needed to run the verification on a thread with a
// stack size of our choosing, see ffi_verifier_thread.go. They are looked
// up through the pact_ffi library, which depends on the C library that
// defines them.
var (
	pthread_attr_init         func(attr unsafe.Pointer) int32
	pthread_attr_setstacksize func(attr unsafe.Pointer, size uintptr) int32
	pthread_attr_destroy      func(attr unsafe.Pointer) int32
	pthread_create            func(thread unsafe.Pointer, attr unsafe.Pointer, start uintptr, arg unsafe.Pointer) int32
	pthread_join              func(thread uintptr, result unsafe.Pointer) int32

	// verifierExecuteAddr is the address of pactffi_verifier_execute, which
	// is the new thread's start routine.
	verifierExecuteAddr uintptr
)

func verifierThreadSymbols() []ffiSymbol {
	return []ffiSymbol{
		{"pthread_attr_init", &pthread_attr_init},
		{"pthread_attr_setstacksize", &pthread_attr_setstacksize},
		{"pthread_attr_destroy", &pthread_attr_destroy},
		{"pthread_create", &pthread_create},
		{"pthread_join", &pthread_join},
		{"pactffi_verifier_execute", &verifierExecuteAddr},
	}
}

// useVerifierThread makes pactffi_verifier_execute run on a thread with a
// large stack. If that thread cannot be created, the call is made on the
// calling thread, as it would be otherwise.
func useVerifierThread() {
	direct := pactffi_verifier_execute

	pactffi_verifier_execute = func(handle unsafe.Pointer) int32 {
		// pthread_attr_t is 64 bytes or fewer in every Linux C library.
		var attr [16]uint64
		if pthread_attr_init(unsafe.Pointer(&attr)) != 0 {
			return direct(handle)
		}

		var thread uintptr // pthread_t is the size of a pointer
		started := pthread_attr_setstacksize(unsafe.Pointer(&attr), verifierStackSize) == 0 &&
			pthread_create(unsafe.Pointer(&thread), unsafe.Pointer(&attr), verifierExecuteAddr, handle) == 0
		pthread_attr_destroy(unsafe.Pointer(&attr))
		if !started {
			return direct(handle)
		}

		// The start routine returns a C int where a pointer is expected, so
		// only the low 32 bits of the result are meaningful.
		var result uintptr
		if errno := pthread_join(thread, unsafe.Pointer(&result)); errno != 0 {
			panic(fmt.Errorf("pact-go: could not wait for the verifier thread: pthread_join failed with %d", errno))
		}

		return int32(result) // #nosec G115 -- truncation to the C int is intended
	}
}
