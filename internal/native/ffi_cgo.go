//go:build cgo && !pact_purego && !pact_goffi

package native

/*
#include <stdlib.h>
#include "pact.h"
*/
import "C"

import "unsafe"

// ffiBackend names the binding in use, see ffi_types.go.
const ffiBackend = "cgo"

// cStringArray converts inputs into a C array of C strings, and returns it
// with a function that frees it. It returns nil for an empty array.
func cStringArray(inputs []string) (unsafe.Pointer, func()) {
	if len(inputs) == 0 {
		return nil, func() {}
	}

	output := make([]*C.char, len(inputs))
	for i, input := range inputs {
		output[i] = C.CString(input)
	}

	freeOutput := func() {
		for _, s := range output {
			C.free(unsafe.Pointer(s))
		}
	}

	return unsafe.Pointer(&output[0]), freeOutput
}

func pactffi_with_multipart_file(interaction uint32, part int32, contentType string, file string, partName string) {
	cContentType := C.CString(contentType)
	defer C.free(unsafe.Pointer(cContentType))
	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))
	cPartName := C.CString(partName)
	defer C.free(unsafe.Pointer(cPartName))
	C.pactffi_with_multipart_file(C.InteractionHandle(interaction), C.int(part), cContentType, cFile, cPartName)
}

func pactffi_version() unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_version())
}

func pactffi_free_string(s unsafe.Pointer) {
	C.pactffi_free_string((*C.char)(s))
}

func pactffi_log_to_stdout(level int32) int32 {
	return int32(C.pactffi_log_to_stdout(C.int(level)))
}

func pactffi_log_to_file(file string, level int32) int32 {
	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))
	return int32(C.pactffi_log_to_file(cFile, C.int(level)))
}

func pactffi_new_pact(consumer string, provider string) uint16 {
	cConsumer := C.CString(consumer)
	defer C.free(unsafe.Pointer(cConsumer))
	cProvider := C.CString(provider)
	defer C.free(unsafe.Pointer(cProvider))
	return uint16(C.pactffi_new_pact(cConsumer, cProvider))
}

func pactffi_with_specification(pact uint16, version int32) bool {
	return bool(C.pactffi_with_specification(C.PactHandle(pact), C.int(version)))
}

func pactffi_with_pact_metadata(pact uint16, namespace string, name string, value string) bool {
	cNamespace := C.CString(namespace)
	defer C.free(unsafe.Pointer(cNamespace))
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_with_pact_metadata(C.PactHandle(pact), cNamespace, cName, cValue))
}

func pactffi_using_plugin(pact uint16, name string, version string) uint32 {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cVersion := C.CString(version)
	defer C.free(unsafe.Pointer(cVersion))
	return uint32(C.pactffi_using_plugin(C.PactHandle(pact), cName, cVersion))
}

func pactffi_cleanup_plugins(pact uint16) {
	C.pactffi_cleanup_plugins(C.PactHandle(pact))
}

func pactffi_new_interaction(pact uint16, description string) uint32 {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	return uint32(C.pactffi_new_interaction(C.PactHandle(pact), cDescription))
}

func pactffi_new_message_interaction(pact uint16, description string) uint32 {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	return uint32(C.pactffi_new_message_interaction(C.PactHandle(pact), cDescription))
}

func pactffi_new_sync_message_interaction(pact uint16, description string) uint32 {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	return uint32(C.pactffi_new_sync_message_interaction(C.PactHandle(pact), cDescription))
}

func pactffi_interaction_contents(interaction uint32, part int32, contentType string, contents string) uint32 {
	cContentType := C.CString(contentType)
	defer C.free(unsafe.Pointer(cContentType))
	cContents := C.CString(contents)
	defer C.free(unsafe.Pointer(cContents))
	return uint32(C.pactffi_interaction_contents(C.InteractionHandle(interaction), C.int(part), cContentType, cContents))
}

func pactffi_upon_receiving(interaction uint32, description string) bool {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	return bool(C.pactffi_upon_receiving(C.InteractionHandle(interaction), cDescription))
}

func pactffi_given(interaction uint32, description string) bool {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	return bool(C.pactffi_given(C.InteractionHandle(interaction), cDescription))
}

func pactffi_given_with_param(interaction uint32, description string, name string, value string) bool {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_given_with_param(C.InteractionHandle(interaction), cDescription, cName, cValue))
}

func pactffi_with_request(interaction uint32, method string, path string) bool {
	cMethod := C.CString(method)
	defer C.free(unsafe.Pointer(cMethod))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	return bool(C.pactffi_with_request(C.InteractionHandle(interaction), cMethod, cPath))
}

func pactffi_with_query_parameter_v2(interaction uint32, name string, index uintptr, value string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_with_query_parameter_v2(C.InteractionHandle(interaction), cName, CUlong(index), cValue))
}

func pactffi_with_header_v2(interaction uint32, part int32, name string, index uintptr, value string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_with_header_v2(C.InteractionHandle(interaction), C.int(part), cName, CUlong(index), cValue))
}

func pactffi_with_body(interaction uint32, part int32, contentType string, body string) bool {
	cContentType := C.CString(contentType)
	defer C.free(unsafe.Pointer(cContentType))
	cBody := C.CString(body)
	defer C.free(unsafe.Pointer(cBody))
	return bool(C.pactffi_with_body(C.InteractionHandle(interaction), C.int(part), cContentType, cBody))
}

func pactffi_with_binary_file(interaction uint32, part int32, contentType string, body unsafe.Pointer, size uintptr) bool {
	cContentType := C.CString(contentType)
	defer C.free(unsafe.Pointer(cContentType))
	return bool(C.pactffi_with_binary_file(C.InteractionHandle(interaction), C.int(part), cContentType, (*C.uint8_t)(body), CUlong(size)))
}

func pactffi_response_status(interaction uint32, status uint16) bool {
	return bool(C.pactffi_response_status(C.InteractionHandle(interaction), C.ushort(status)))
}

func pactffi_add_interaction_reference(interaction uint32, group string, name string, value string) bool {
	cGroup := C.CString(group)
	defer C.free(unsafe.Pointer(cGroup))
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_add_interaction_reference(C.InteractionHandle(interaction), cGroup, cName, cValue))
}

func pactffi_with_metadata(interaction uint32, key string, value string, part int32) bool {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return bool(C.pactffi_with_metadata(C.InteractionHandle(interaction), cKey, cValue, C.int(part)))
}

func pactffi_create_mock_server_for_transport(pact uint16, addr string, port uint16, transport string, config *byte) int32 {
	cAddr := C.CString(addr)
	defer C.free(unsafe.Pointer(cAddr))
	cTransport := C.CString(transport)
	defer C.free(unsafe.Pointer(cTransport))
	return int32(C.pactffi_create_mock_server_for_transport(C.PactHandle(pact), cAddr, C.uint16_t(port), cTransport, (*C.char)(unsafe.Pointer(config))))
}

func pactffi_mock_server_mismatches(port int32) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_mock_server_mismatches(C.int32_t(port)))
}

func pactffi_mock_server_matched(port int32) bool {
	return bool(C.pactffi_mock_server_matched(C.int32_t(port)))
}

func pactffi_cleanup_mock_server(port int32) bool {
	return bool(C.pactffi_cleanup_mock_server(C.int32_t(port)))
}

func pactffi_write_pact_file(port int32, dir string, overwrite bool) int32 {
	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))
	return int32(C.pactffi_write_pact_file(C.int32_t(port), cDir, C.bool(overwrite)))
}

func pactffi_get_tls_ca_certificate() unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_get_tls_ca_certificate())
}

func pactffi_new_message_pact(consumer string, provider string) uint16 {
	cConsumer := C.CString(consumer)
	defer C.free(unsafe.Pointer(cConsumer))
	cProvider := C.CString(provider)
	defer C.free(unsafe.Pointer(cProvider))
	return uint16(C.pactffi_new_message_pact(cConsumer, cProvider))
}

func pactffi_with_message_pact_metadata(pact uint16, namespace string, name string, value string) {
	cNamespace := C.CString(namespace)
	defer C.free(unsafe.Pointer(cNamespace))
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.pactffi_with_message_pact_metadata(C.MessagePactHandle(pact), cNamespace, cName, cValue)
}

func pactffi_message_expects_to_receive(message uint32, description string) {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	C.pactffi_message_expects_to_receive(C.MessageHandle(message), cDescription)
}

func pactffi_message_with_metadata(message uint32, key string, value string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.pactffi_message_with_metadata(C.MessageHandle(message), cKey, cValue)
}

func pactffi_write_message_pact_file(pact uint16, dir string, overwrite bool) int32 {
	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))
	return int32(C.pactffi_write_message_pact_file(C.MessagePactHandle(pact), cDir, C.bool(overwrite)))
}

func pactffi_pact_handle_get_message_iter(pact uint16) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_pact_handle_get_message_iter(C.PactHandle(pact)))
}

func pactffi_pact_message_iter_next(iter unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_pact_message_iter_next((*C.struct_PactMessageIterator)(iter)))
}

func pactffi_pact_message_iter_delete(iter unsafe.Pointer) {
	C.pactffi_pact_message_iter_delete((*C.struct_PactMessageIterator)(iter))
}

func pactffi_message_get_contents_length(message unsafe.Pointer) uintptr {
	return uintptr(C.pactffi_message_get_contents_length((*C.struct_Message)(message)))
}

func pactffi_message_get_contents_bin(message unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_message_get_contents_bin((*C.struct_Message)(message)))
}

func pactffi_pact_handle_get_sync_message_iter(pact uint16) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_pact_handle_get_sync_message_iter(C.PactHandle(pact)))
}

func pactffi_pact_sync_message_iter_next(iter unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_pact_sync_message_iter_next((*C.struct_PactSyncMessageIterator)(iter)))
}

func pactffi_pact_sync_message_iter_delete(iter unsafe.Pointer) {
	C.pactffi_pact_sync_message_iter_delete((*C.struct_PactSyncMessageIterator)(iter))
}

func pactffi_sync_message_get_request_contents_length(message unsafe.Pointer) uintptr {
	return uintptr(C.pactffi_sync_message_get_request_contents_length((*C.struct_SynchronousMessage)(message)))
}

func pactffi_sync_message_get_request_contents_bin(message unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_sync_message_get_request_contents_bin((*C.struct_SynchronousMessage)(message)))
}

func pactffi_sync_message_get_response_contents_length(message unsafe.Pointer, index uintptr) uintptr {
	return uintptr(C.pactffi_sync_message_get_response_contents_length((*C.struct_SynchronousMessage)(message), CUlong(index)))
}

func pactffi_sync_message_get_response_contents_bin(message unsafe.Pointer, index uintptr) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_sync_message_get_response_contents_bin((*C.struct_SynchronousMessage)(message), CUlong(index)))
}

func pactffi_verifier_new_for_application(name string, version string) unsafe.Pointer {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cVersion := C.CString(version)
	defer C.free(unsafe.Pointer(cVersion))
	return unsafe.Pointer(C.pactffi_verifier_new_for_application(cName, cVersion))
}

func pactffi_verifier_shutdown(handle unsafe.Pointer) {
	C.pactffi_verifier_shutdown((*C.struct_VerifierHandle)(handle))
}

func pactffi_verifier_set_provider_info(handle unsafe.Pointer, name string, scheme string, host string, port uint16, path string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.pactffi_verifier_set_provider_info((*C.struct_VerifierHandle)(handle), cName, cScheme, cHost, C.ushort(port), cPath)
}

func pactffi_verifier_add_provider_transport(handle unsafe.Pointer, protocol string, port uint16, path string, scheme string) {
	cProtocol := C.CString(protocol)
	defer C.free(unsafe.Pointer(cProtocol))
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))
	C.pactffi_verifier_add_provider_transport((*C.struct_VerifierHandle)(handle), cProtocol, C.ushort(port), cPath, cScheme)
}

func pactffi_verifier_set_filter_info(handle unsafe.Pointer, description string, state string, noState uint8) {
	cDescription := C.CString(description)
	defer C.free(unsafe.Pointer(cDescription))
	cState := C.CString(state)
	defer C.free(unsafe.Pointer(cState))
	C.pactffi_verifier_set_filter_info((*C.struct_VerifierHandle)(handle), cDescription, cState, C.uchar(noState))
}

func pactffi_verifier_set_provider_state(handle unsafe.Pointer, sourceURL string, teardown uint8, body uint8) {
	cSourceURL := C.CString(sourceURL)
	defer C.free(unsafe.Pointer(cSourceURL))
	C.pactffi_verifier_set_provider_state((*C.struct_VerifierHandle)(handle), cSourceURL, C.uchar(teardown), C.uchar(body))
}

func pactffi_verifier_set_verification_options(handle unsafe.Pointer, disableSSLVerification uint8, requestTimeout uint64) int32 {
	return int32(C.pactffi_verifier_set_verification_options((*C.struct_VerifierHandle)(handle), C.uchar(disableSSLVerification), C.ulong(requestTimeout)))
}

func pactffi_verifier_set_consumer_filters(handle unsafe.Pointer, filters unsafe.Pointer, filtersLen uint16) {
	C.pactffi_verifier_set_consumer_filters((*C.struct_VerifierHandle)(handle), (**C.char)(filters), C.ushort(filtersLen))
}

func pactffi_verifier_add_custom_header(handle unsafe.Pointer, name string, value string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.pactffi_verifier_add_custom_header((*C.struct_VerifierHandle)(handle), cName, cValue)
}

func pactffi_verifier_add_file_source(handle unsafe.Pointer, file string) {
	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))
	C.pactffi_verifier_add_file_source((*C.struct_VerifierHandle)(handle), cFile)
}

func pactffi_verifier_add_directory_source(handle unsafe.Pointer, directory string) {
	cDirectory := C.CString(directory)
	defer C.free(unsafe.Pointer(cDirectory))
	C.pactffi_verifier_add_directory_source((*C.struct_VerifierHandle)(handle), cDirectory)
}

func pactffi_verifier_url_source(handle unsafe.Pointer, sourceURL string, username string, password string, token string) {
	cSourceURL := C.CString(sourceURL)
	defer C.free(unsafe.Pointer(cSourceURL))
	cUsername := C.CString(username)
	defer C.free(unsafe.Pointer(cUsername))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	cToken := C.CString(token)
	defer C.free(unsafe.Pointer(cToken))
	C.pactffi_verifier_url_source((*C.struct_VerifierHandle)(handle), cSourceURL, cUsername, cPassword, cToken)
}

func pactffi_verifier_broker_source_with_selectors(handle unsafe.Pointer, sourceURL string, username string, password string, token string, enablePending uint8, includeWipPactsSince string, providerTags unsafe.Pointer, providerTagsLen uint16, providerBranch string, selectors unsafe.Pointer, selectorsLen uint16, consumerVersionTags unsafe.Pointer, consumerVersionTagsLen uint16) int32 {
	cSourceURL := C.CString(sourceURL)
	defer C.free(unsafe.Pointer(cSourceURL))
	cUsername := C.CString(username)
	defer C.free(unsafe.Pointer(cUsername))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	cToken := C.CString(token)
	defer C.free(unsafe.Pointer(cToken))
	cIncludeWipPactsSince := C.CString(includeWipPactsSince)
	defer C.free(unsafe.Pointer(cIncludeWipPactsSince))
	cProviderBranch := C.CString(providerBranch)
	defer C.free(unsafe.Pointer(cProviderBranch))
	return int32(C.pactffi_verifier_broker_source_with_selectors((*C.struct_VerifierHandle)(handle), cSourceURL, cUsername, cPassword, cToken, C.uchar(enablePending), cIncludeWipPactsSince, (**C.char)(providerTags), C.ushort(providerTagsLen), cProviderBranch, (**C.char)(selectors), C.ushort(selectorsLen), (**C.char)(consumerVersionTags), C.ushort(consumerVersionTagsLen)))
}

func pactffi_verifier_set_publish_options(handle unsafe.Pointer, providerVersion string, buildURL string, providerTags unsafe.Pointer, providerTagsLen uint16, providerBranch string) int32 {
	cProviderVersion := C.CString(providerVersion)
	defer C.free(unsafe.Pointer(cProviderVersion))
	cBuildURL := C.CString(buildURL)
	defer C.free(unsafe.Pointer(cBuildURL))
	cProviderBranch := C.CString(providerBranch)
	defer C.free(unsafe.Pointer(cProviderBranch))
	return int32(C.pactffi_verifier_set_publish_options((*C.struct_VerifierHandle)(handle), cProviderVersion, cBuildURL, (**C.char)(providerTags), C.ushort(providerTagsLen), cProviderBranch))
}

func pactffi_verifier_execute(handle unsafe.Pointer) int32 {
	return int32(C.pactffi_verifier_execute((*C.struct_VerifierHandle)(handle)))
}

func pactffi_verifier_set_no_pacts_is_error(handle unsafe.Pointer, isError uint8) int32 {
	return int32(C.pactffi_verifier_set_no_pacts_is_error((*C.struct_VerifierHandle)(handle), C.uchar(isError)))
}

func pactffi_verifier_set_coloured_output(handle unsafe.Pointer, colouredOutput uint8) int32 {
	return int32(C.pactffi_verifier_set_coloured_output((*C.struct_VerifierHandle)(handle), C.uchar(colouredOutput)))
}

func pactffi_verifier_json(handle unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.pactffi_verifier_json((*C.struct_VerifierHandle)(handle)))
}
