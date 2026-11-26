//go:build (!cgo || pact_purego || pact_goffi) && !linux

package native

// Only Linux runs the verification on a thread of its own, see
// ffi_verifier_thread.go.

func verifierThreadSymbols() []ffiSymbol { return nil }

func useVerifierThread() {}
