// Package native contains the c bindings into the Pact Reference types.
package native

/*
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/libs/macos-aarch64 -L/tmp -L/usr/local/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,${SRCDIR}/libs/macos-aarch64 -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/libs/macos-x86_64 -L/tmp -L/usr/local/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,${SRCDIR}/libs/macos-x86_64 -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/libs/windows-x86_64 -lpact_ffi
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/libs/linux-aarch64 -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,${SRCDIR}/libs/linux-aarch64 -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux-x86_64 -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,${SRCDIR}/libs/linux-x86_64 -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
*/
import "C"

// in order to correctly link to our embedded files

// install_name_tool -id "@rpath/libpact_ffi.dylib" internal/native/libs/macos-aarch64/libpact_ffi.dylib
// install_name_tool -id "@rpath/libpact_ffi.dylib" internal/native/libs/macos-x86_64/libpact_ffi.dylib