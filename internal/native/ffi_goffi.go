//go:build pact_goffi

package native

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/go-webgpu/goffi/ffi"
	"github.com/go-webgpu/goffi/types"
)

// ffiBackend names the binding in use, see ffi_types.go.
const ffiBackend = "goffi"

// bindLibrary loads the pact_ffi library at path and binds every symbol
// to its Go func variable.
func bindLibrary(path string, symbols []ffiSymbol) error {
	lib, err := ffi.LoadLibrary(path)
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}

	// Resolve and prepare everything before binding anything, so that a
	// failure leaves no function half bound.
	bound := make([]reflect.Value, len(symbols))
	for i, s := range symbols {
		addr, err := ffi.GetSymbol(lib, s.name)
		if err != nil {
			return fmt.Errorf("symbol %s: %w", s.name, err)
		}

		bound[i], err = goffiFunc(reflect.TypeOf(s.fptr).Elem(), addr)
		if err != nil {
			return fmt.Errorf("symbol %s: %w", s.name, err)
		}
	}
	for i, s := range symbols {
		reflect.ValueOf(s.fptr).Elem().Set(bound[i])
	}

	return nil
}

// goffiType maps a Go parameter or result type to its C type, following
// the table in ffi_types.go.
func goffiType(t reflect.Type) (*types.TypeDescriptor, error) {
	switch t.Kind() { //nolint:exhaustive // only the kinds in the ffi_types.go table are bound; goffiType rejects the rest
	case reflect.String, reflect.UnsafePointer, reflect.Pointer:
		return types.PointerTypeDescriptor, nil
	case reflect.Bool, reflect.Uint8:
		return types.UInt8TypeDescriptor, nil
	case reflect.Uint16:
		return types.UInt16TypeDescriptor, nil
	case reflect.Uint32:
		return types.UInt32TypeDescriptor, nil
	case reflect.Int32:
		return types.SInt32TypeDescriptor, nil
	case reflect.Uint64, reflect.Uintptr:
		return types.UInt64TypeDescriptor, nil
	default:
		return nil, fmt.Errorf("unsupported type %s", t)
	}
}

// goffiFunc returns a Go func of type t that calls the C function at addr.
func goffiFunc(t reflect.Type, addr unsafe.Pointer) (reflect.Value, error) {
	argTypes := make([]*types.TypeDescriptor, t.NumIn())
	for i := range argTypes {
		var err error
		argTypes[i], err = goffiType(t.In(i))
		if err != nil {
			return reflect.Value{}, fmt.Errorf("argument %d: %w", i, err)
		}
	}

	retType := types.VoidTypeDescriptor
	if t.NumOut() == 1 {
		var err error
		retType, err = goffiType(t.Out(0))
		if err != nil {
			return reflect.Value{}, fmt.Errorf("result: %w", err)
		}
	}

	cif := &types.CallInterface{}
	err := ffi.PrepareCallInterface(cif, types.DefaultCall, retType, argTypes)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("prepare call: %w", err)
	}

	return reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
		// goffi takes a pointer to each argument. Pointers and scalars are
		// kept in separate slots so that the garbage collector sees the
		// pointers; scalars are little-endian, so the C type can be read
		// from the start of a 64 bit slot.
		ptrs := make([]unsafe.Pointer, len(args))
		vals := make([]uint64, len(args))
		avalue := make([]unsafe.Pointer, len(args))
		for i, arg := range args {
			avalue[i] = unsafe.Pointer(&vals[i])

			switch arg.Kind() { //nolint:exhaustive // only the kinds in the ffi_types.go table are bound; goffiType rejects the rest
			case reflect.String:
				ptrs[i] = unsafe.Pointer(cstr(arg.String()))
				avalue[i] = unsafe.Pointer(&ptrs[i])
			case reflect.UnsafePointer, reflect.Pointer:
				ptrs[i] = arg.UnsafePointer()
				avalue[i] = unsafe.Pointer(&ptrs[i])
			case reflect.Bool:
				if arg.Bool() {
					vals[i] = 1
				}
			case reflect.Int32:
				vals[i] = uint64(arg.Int()) // #nosec G115 -- sign extension is intended
			default:
				vals[i] = arg.Uint()
			}
		}

		var (
			retVal uint64
			retPtr unsafe.Pointer
			rvalue unsafe.Pointer
		)
		isPtr := t.NumOut() == 1 && t.Out(0).Kind() == reflect.UnsafePointer
		switch {
		case isPtr:
			rvalue = unsafe.Pointer(&retPtr)
		case t.NumOut() == 1:
			rvalue = unsafe.Pointer(&retVal)
		}

		_, err := ffi.CallFunction(cif, addr, rvalue, avalue)
		runtime.KeepAlive(ptrs)
		runtime.KeepAlive(vals)
		if err != nil {
			panic(fmt.Errorf("pact-go: goffi call failed: %w", err))
		}

		if t.NumOut() == 0 {
			return nil
		}

		out := reflect.New(t.Out(0)).Elem()
		switch out.Kind() { //nolint:exhaustive // only the kinds in the ffi_types.go table are bound; goffiType rejects the rest
		case reflect.UnsafePointer:
			out.SetPointer(retPtr)
		case reflect.Bool:
			out.SetBool(uint8(retVal) != 0)
		case reflect.Int32:
			out.SetInt(int64(int32(retVal))) // #nosec G115 -- truncation to the C type is intended
		case reflect.Uint16:
			out.SetUint(retVal & math.MaxUint16)
		case reflect.Uint32:
			out.SetUint(retVal & math.MaxUint32)
		default:
			out.SetUint(retVal)
		}

		return []reflect.Value{out}
	}), nil
}
