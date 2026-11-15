//go:build !cgo || pact_purego || pact_goffi

package native

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"unsafe"
)

// The pact_ffi functions, bound at runtime by the purego or goffi backend
// (see ffi_types.go). Only the functions this package calls are bound.
var (
	pactffi_version                                   func() unsafe.Pointer
	pactffi_free_string                               func(s unsafe.Pointer)
	pactffi_log_to_stdout                             func(level int32) int32
	pactffi_log_to_file                               func(file string, level int32) int32
	pactffi_new_pact                                  func(consumer string, provider string) uint16
	pactffi_with_specification                        func(pact uint16, version int32) bool
	pactffi_with_pact_metadata                        func(pact uint16, namespace string, name string, value string) bool
	pactffi_using_plugin                              func(pact uint16, name string, version string) uint32
	pactffi_cleanup_plugins                           func(pact uint16)
	pactffi_new_interaction                           func(pact uint16, description string) uint32
	pactffi_new_message_interaction                   func(pact uint16, description string) uint32
	pactffi_new_sync_message_interaction              func(pact uint16, description string) uint32
	pactffi_interaction_contents                      func(interaction uint32, part int32, contentType string, contents string) uint32
	pactffi_upon_receiving                            func(interaction uint32, description string) bool
	pactffi_given                                     func(interaction uint32, description string) bool
	pactffi_given_with_param                          func(interaction uint32, description string, name string, value string) bool
	pactffi_with_request                              func(interaction uint32, method string, path string) bool
	pactffi_with_query_parameter_v2                   func(interaction uint32, name string, index uintptr, value string) bool
	pactffi_with_header_v2                            func(interaction uint32, part int32, name string, index uintptr, value string) bool
	pactffi_with_body                                 func(interaction uint32, part int32, contentType string, body string) bool
	pactffi_with_binary_file                          func(interaction uint32, part int32, contentType string, body unsafe.Pointer, size uintptr) bool
	pactffi_response_status                           func(interaction uint32, status uint16) bool
	pactffi_add_interaction_reference                 func(interaction uint32, group string, name string, value string) bool
	pactffi_with_metadata                             func(interaction uint32, key string, value string, part int32) bool
	pactffi_create_mock_server_for_transport          func(pact uint16, addr string, port uint16, transport string, config *byte) int32
	pactffi_mock_server_mismatches                    func(port int32) unsafe.Pointer
	pactffi_mock_server_matched                       func(port int32) bool
	pactffi_cleanup_mock_server                       func(port int32) bool
	pactffi_write_pact_file                           func(port int32, dir string, overwrite bool) int32
	pactffi_get_tls_ca_certificate                    func() unsafe.Pointer
	pactffi_new_message_pact                          func(consumer string, provider string) uint16
	pactffi_with_message_pact_metadata                func(pact uint16, namespace string, name string, value string)
	pactffi_message_expects_to_receive                func(message uint32, description string)
	pactffi_message_with_metadata                     func(message uint32, key string, value string)
	pactffi_write_message_pact_file                   func(pact uint16, dir string, overwrite bool) int32
	pactffi_pact_handle_get_message_iter              func(pact uint16) unsafe.Pointer
	pactffi_pact_message_iter_next                    func(iter unsafe.Pointer) unsafe.Pointer
	pactffi_pact_message_iter_delete                  func(iter unsafe.Pointer)
	pactffi_message_get_contents_length               func(message unsafe.Pointer) uintptr
	pactffi_message_get_contents_bin                  func(message unsafe.Pointer) unsafe.Pointer
	pactffi_pact_handle_get_sync_message_iter         func(pact uint16) unsafe.Pointer
	pactffi_pact_sync_message_iter_next               func(iter unsafe.Pointer) unsafe.Pointer
	pactffi_pact_sync_message_iter_delete             func(iter unsafe.Pointer)
	pactffi_sync_message_get_request_contents_length  func(message unsafe.Pointer) uintptr
	pactffi_sync_message_get_request_contents_bin     func(message unsafe.Pointer) unsafe.Pointer
	pactffi_sync_message_get_response_contents_length func(message unsafe.Pointer, index uintptr) uintptr
	pactffi_sync_message_get_response_contents_bin    func(message unsafe.Pointer, index uintptr) unsafe.Pointer
	pactffi_verifier_new_for_application              func(name string, version string) unsafe.Pointer
	pactffi_verifier_shutdown                         func(handle unsafe.Pointer)
	pactffi_verifier_set_provider_info                func(handle unsafe.Pointer, name string, scheme string, host string, port uint16, path string)
	pactffi_verifier_add_provider_transport           func(handle unsafe.Pointer, protocol string, port uint16, path string, scheme string)
	pactffi_verifier_set_filter_info                  func(handle unsafe.Pointer, description string, state string, noState uint8)
	pactffi_verifier_set_provider_state               func(handle unsafe.Pointer, sourceURL string, teardown uint8, body uint8)
	pactffi_verifier_set_verification_options         func(handle unsafe.Pointer, disableSSLVerification uint8, requestTimeout uint64) int32
	pactffi_verifier_set_consumer_filters             func(handle unsafe.Pointer, filters unsafe.Pointer, filtersLen uint16)
	pactffi_verifier_add_custom_header                func(handle unsafe.Pointer, name string, value string)
	pactffi_verifier_add_file_source                  func(handle unsafe.Pointer, file string)
	pactffi_verifier_add_directory_source             func(handle unsafe.Pointer, directory string)
	pactffi_verifier_url_source                       func(handle unsafe.Pointer, sourceURL string, username string, password string, token string)
	pactffi_verifier_broker_source_with_selectors     func(handle unsafe.Pointer, sourceURL string, username string, password string, token string, enablePending uint8, includeWipPactsSince string, providerTags unsafe.Pointer, providerTagsLen uint16, providerBranch string, selectors unsafe.Pointer, selectorsLen uint16, consumerVersionTags unsafe.Pointer, consumerVersionTagsLen uint16) int32
	pactffi_verifier_set_publish_options              func(handle unsafe.Pointer, providerVersion string, buildURL string, providerTags unsafe.Pointer, providerTagsLen uint16, providerBranch string) int32
	pactffi_verifier_execute                          func(handle unsafe.Pointer) int32
	pactffi_verifier_set_no_pacts_is_error            func(handle unsafe.Pointer, isError uint8) int32
	pactffi_verifier_set_coloured_output              func(handle unsafe.Pointer, colouredOutput uint8) int32
	pactffi_verifier_json                             func(handle unsafe.Pointer) unsafe.Pointer

	// pactffi_with_multipart_file returns a 16 byte struct by value, which
	// is not read. On the System V and AAPCS64 ABIs it comes back in
	// registers and can be ignored; on Windows x64 the caller must pass a
	// buffer for it as a hidden first argument.
	withMultipartFileRegs func(interaction uint32, part int32, contentType string, file string, partName string)
	withMultipartFileMem  func(result unsafe.Pointer, interaction uint32, part int32, contentType string, file string, partName string) unsafe.Pointer
)

func pactffi_with_multipart_file(interaction uint32, part int32, contentType string, file string, partName string) {
	if runtime.GOOS != goosWindows {
		withMultipartFileRegs(interaction, part, contentType, file, partName)

		return
	}

	var result [2]uintptr
	withMultipartFileMem(unsafe.Pointer(&result), interaction, part, contentType, file, partName)
}

// ffiSymbol pairs the name of a pact_ffi function with a pointer to the Go
// func variable it is bound to. If fptr is a *uintptr instead, the symbol
// is not bound: its address is stored there.
type ffiSymbol struct {
	name string
	fptr any
}

func ffiSymbols() []ffiSymbol {
	symbols := []ffiSymbol{
		{"pactffi_version", &pactffi_version},
		{"pactffi_free_string", &pactffi_free_string},
		{"pactffi_log_to_stdout", &pactffi_log_to_stdout},
		{"pactffi_log_to_file", &pactffi_log_to_file},
		{"pactffi_new_pact", &pactffi_new_pact},
		{"pactffi_with_specification", &pactffi_with_specification},
		{"pactffi_with_pact_metadata", &pactffi_with_pact_metadata},
		{"pactffi_using_plugin", &pactffi_using_plugin},
		{"pactffi_cleanup_plugins", &pactffi_cleanup_plugins},
		{"pactffi_new_interaction", &pactffi_new_interaction},
		{"pactffi_new_message_interaction", &pactffi_new_message_interaction},
		{"pactffi_new_sync_message_interaction", &pactffi_new_sync_message_interaction},
		{"pactffi_interaction_contents", &pactffi_interaction_contents},
		{"pactffi_upon_receiving", &pactffi_upon_receiving},
		{"pactffi_given", &pactffi_given},
		{"pactffi_given_with_param", &pactffi_given_with_param},
		{"pactffi_with_request", &pactffi_with_request},
		{"pactffi_with_query_parameter_v2", &pactffi_with_query_parameter_v2},
		{"pactffi_with_header_v2", &pactffi_with_header_v2},
		{"pactffi_with_body", &pactffi_with_body},
		{"pactffi_with_binary_file", &pactffi_with_binary_file},
		{"pactffi_response_status", &pactffi_response_status},
		{"pactffi_add_interaction_reference", &pactffi_add_interaction_reference},
		{"pactffi_with_metadata", &pactffi_with_metadata},
		{"pactffi_create_mock_server_for_transport", &pactffi_create_mock_server_for_transport},
		{"pactffi_mock_server_mismatches", &pactffi_mock_server_mismatches},
		{"pactffi_mock_server_matched", &pactffi_mock_server_matched},
		{"pactffi_cleanup_mock_server", &pactffi_cleanup_mock_server},
		{"pactffi_write_pact_file", &pactffi_write_pact_file},
		{"pactffi_get_tls_ca_certificate", &pactffi_get_tls_ca_certificate},
		{"pactffi_new_message_pact", &pactffi_new_message_pact},
		{"pactffi_with_message_pact_metadata", &pactffi_with_message_pact_metadata},
		{"pactffi_message_expects_to_receive", &pactffi_message_expects_to_receive},
		{"pactffi_message_with_metadata", &pactffi_message_with_metadata},
		{"pactffi_write_message_pact_file", &pactffi_write_message_pact_file},
		{"pactffi_pact_handle_get_message_iter", &pactffi_pact_handle_get_message_iter},
		{"pactffi_pact_message_iter_next", &pactffi_pact_message_iter_next},
		{"pactffi_pact_message_iter_delete", &pactffi_pact_message_iter_delete},
		{"pactffi_message_get_contents_length", &pactffi_message_get_contents_length},
		{"pactffi_message_get_contents_bin", &pactffi_message_get_contents_bin},
		{"pactffi_pact_handle_get_sync_message_iter", &pactffi_pact_handle_get_sync_message_iter},
		{"pactffi_pact_sync_message_iter_next", &pactffi_pact_sync_message_iter_next},
		{"pactffi_pact_sync_message_iter_delete", &pactffi_pact_sync_message_iter_delete},
		{"pactffi_sync_message_get_request_contents_length", &pactffi_sync_message_get_request_contents_length},
		{"pactffi_sync_message_get_request_contents_bin", &pactffi_sync_message_get_request_contents_bin},
		{"pactffi_sync_message_get_response_contents_length", &pactffi_sync_message_get_response_contents_length},
		{"pactffi_sync_message_get_response_contents_bin", &pactffi_sync_message_get_response_contents_bin},
		{"pactffi_verifier_new_for_application", &pactffi_verifier_new_for_application},
		{"pactffi_verifier_shutdown", &pactffi_verifier_shutdown},
		{"pactffi_verifier_set_provider_info", &pactffi_verifier_set_provider_info},
		{"pactffi_verifier_add_provider_transport", &pactffi_verifier_add_provider_transport},
		{"pactffi_verifier_set_filter_info", &pactffi_verifier_set_filter_info},
		{"pactffi_verifier_set_provider_state", &pactffi_verifier_set_provider_state},
		{"pactffi_verifier_set_verification_options", &pactffi_verifier_set_verification_options},
		{"pactffi_verifier_set_consumer_filters", &pactffi_verifier_set_consumer_filters},
		{"pactffi_verifier_add_custom_header", &pactffi_verifier_add_custom_header},
		{"pactffi_verifier_add_file_source", &pactffi_verifier_add_file_source},
		{"pactffi_verifier_add_directory_source", &pactffi_verifier_add_directory_source},
		{"pactffi_verifier_url_source", &pactffi_verifier_url_source},
		{"pactffi_verifier_broker_source_with_selectors", &pactffi_verifier_broker_source_with_selectors},
		{"pactffi_verifier_set_publish_options", &pactffi_verifier_set_publish_options},
		{"pactffi_verifier_execute", &pactffi_verifier_execute},
		{"pactffi_verifier_set_no_pacts_is_error", &pactffi_verifier_set_no_pacts_is_error},
		{"pactffi_verifier_set_coloured_output", &pactffi_verifier_set_coloured_output},
		{"pactffi_verifier_json", &pactffi_verifier_json},
	}

	symbols = append(symbols, verifierThreadSymbols()...)

	if runtime.GOOS == goosWindows {
		return append(symbols, ffiSymbol{"pactffi_with_multipart_file", &withMultipartFileMem})
	}

	return append(symbols, ffiSymbol{"pactffi_with_multipart_file", &withMultipartFileRegs})
}

// libEnvVar names a directory to load the pact_ffi library from, ahead of
// the default locations.
// goosWindows is the runtime.GOOS value on Windows.
const goosWindows = "windows"

const libEnvVar = "PACT_LD_LIBRARY_PATH"

// libDownloadEnvVar names the directory `pact-go install` downloads the
// library to, see the installer package.
const libDownloadEnvVar = "PACT_GO_LIB_DOWNLOAD_PATH"

// errLibNotLoaded is the reason the pact_ffi library is unusable, or nil
// if it was loaded and every function was bound.
var errLibNotLoaded error

// LibraryLoadError returns the reason the pact_ffi library could not be
// loaded at startup, or nil if it is ready for use. With the cgo backend
// the library is linked at build time and this is always nil.
func LibraryLoadError() error {
	return errLibNotLoaded
}

func libFileName() string {
	switch runtime.GOOS {
	case goosWindows:
		return "pact_ffi.dll"
	case "darwin":
		return "libpact_ffi.dylib"
	default:
		return "libpact_ffi.so"
	}
}

// libCandidates lists where to look for the pact_ffi library, in order:
// $PACT_LD_LIBRARY_PATH, $PACT_GO_LIB_DOWNLOAD_PATH, the directories the cgo backend links against
// (see lib.go), and finally the bare file name, which leaves the search
// to the system loader (LD_LIBRARY_PATH, DYLD_LIBRARY_PATH, PATH, ...).
func libCandidates() []string {
	name := libFileName()

	var candidates []string
	for _, env := range []string{libEnvVar, libDownloadEnvVar} {
		if dir := os.Getenv(env); dir != "" {
			candidates = append(candidates, filepath.Join(dir, name))
		}
	}
	if runtime.GOOS != goosWindows {
		for _, dir := range []string{"/tmp", "/opt/pact/lib", "/usr/local/lib"} {
			candidates = append(candidates, filepath.Join(dir, name))
		}
	}

	return append(candidates, name)
}

// The package must be importable without the library: `pact-go install`,
// which downloads it, imports this package. So a failure to load is not
// fatal here; instead every pact_ffi function panics with the reason when
// it is first called.
func init() {
	symbols := ffiSymbols()

	errLibNotLoaded = loadLibrary(symbols)
	if errLibNotLoaded == nil {
		useVerifierThread()

		return
	}

	errLibNotLoaded = fmt.Errorf("pact-go: the pact_ffi library is not available to the %s backend (run `pact-go install`, or set %s to the directory containing %s): %w", ffiBackend, libEnvVar, libFileName(), errLibNotLoaded)
	for _, s := range symbols {
		fn := reflect.ValueOf(s.fptr).Elem()
		if fn.Kind() != reflect.Func {
			continue
		}
		fn.Set(reflect.MakeFunc(fn.Type(), func([]reflect.Value) []reflect.Value {
			panic(errLibNotLoaded)
		}))
	}
}

// loadLibrary opens the first candidate library that exists and binds
// every symbol from it.
func loadLibrary(symbols []ffiSymbol) error {
	var errs []error
	for _, path := range libCandidates() {
		if filepath.IsAbs(path) {
			_, err := os.Stat(path)
			if err != nil {
				continue
			}
		}

		err := bindLibrary(path, symbols)
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", path, err))
	}

	return errors.Join(errs...)
}

// cStringArray converts inputs into a C array of C strings, and returns it
// with a function that releases it. It returns nil for an empty array.
func cStringArray(inputs []string) (unsafe.Pointer, func()) {
	if len(inputs) == 0 {
		return nil, func() {}
	}

	output := make([]*byte, len(inputs))
	for i, input := range inputs {
		output[i] = cstr(input)
	}

	return unsafe.Pointer(&output[0]), func() { runtime.KeepAlive(output) }
}
