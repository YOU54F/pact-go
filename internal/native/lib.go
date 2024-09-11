// Package native contains the c bindings into the Pact Reference types.
package native

/*
#cgo darwin,arm64 LDFLAGS: -L/tmp -L/usr/local/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi, -L${SRCDIR}/internal/native/libs/macos-aarch64
#cgo darwin,amd64 LDFLAGS: -L/tmp -L/usr/local/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi, -L${SRCDIR}/internal/native/libs/macos-x86_64
#cgo windows,amd64 LDFLAGS: -lpact_ffi
#cgo linux,amd64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi, -L${SRCDIR}/internal/native/libs/linux-aarch64
#cgo linux,arm64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi, -L${SRCDIR}/internal/native/libs/linux-x86_64
*/
import "C"
