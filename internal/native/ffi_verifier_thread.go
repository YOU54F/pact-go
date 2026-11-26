package native

// Why the verification runs on a thread of its own on Linux.
//
// pactffi_verifier_execute drives the whole verification on the thread that
// calls it, and needs more stack than that thread may have. Go creates its
// threads with the default pthread attributes, so their stack is the
// platform's default: 8 MiB with glibc, but 128 KiB with musl (Alpine),
// which the verification overflows, crashing the process with SIGSEGV. The
// crash was intermittent, as a call that happened to land on the main thread
// had the main thread's 8 MiB.
//
// So on Linux every backend calls pactffi_verifier_execute on a thread it
// creates with an 8 MiB stack, and waits for it: the cgo backend in the C
// preamble of ffi_cgo.go, the purego and goffi backends in
// ffi_verifier_thread_linux.go. With glibc this changes nothing but the
// thread, and other platforms call the function directly, as before.
//
// Versions of pact_ffi that do this themselves make it redundant, but
// harmless.
