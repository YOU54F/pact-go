package native

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

type interactionPart int

// Which half of an interaction pactffi_interaction_contents (and the
// int32(part) calls throughout this file) applies to.
const (
	INTERACTION_PART_REQUEST interactionPart = iota
	INTERACTION_PART_RESPONSE
)

// Generic pass/fail result codes; currently unused in this package.
const (
	RESULT_OK interactionPart = iota
	RESULT_FAILED
)

type specificationVersion int

// Pact specification versions accepted by pactffi_with_specification.
const (
	SPECIFICATION_VERSION_UNKNOWN specificationVersion = iota
	SPECIFICATION_VERSION_V1
	SPECIFICATION_VERSION_V1_1
	SPECIFICATION_VERSION_V2
	SPECIFICATION_VERSION_V3
	SPECIFICATION_VERSION_V4
)

type logLevel int

// Log levels accepted by the pact_ffi logging functions, from least to
// most verbose.
const (
	LOG_LEVEL_OFF logLevel = iota
	LOG_LEVEL_ERROR
	LOG_LEVEL_WARN
	LOG_LEVEL_INFO
	LOG_LEVEL_DEBUG
	LOG_LEVEL_TRACE
)

var logLevelStringToInt = map[string]logLevel{
	"OFF":   LOG_LEVEL_OFF,
	"ERROR": LOG_LEVEL_ERROR,
	"WARN":  LOG_LEVEL_WARN,
	"INFO":  LOG_LEVEL_INFO,
	"DEBUG": LOG_LEVEL_DEBUG,
	"TRACE": LOG_LEVEL_TRACE,
}

// Pact is a Go representation of the PactHandle struct.
type Pact struct {
	handle uint16
}

// Interaction is a Go representation of the InteractionHandle struct.
type Interaction struct {
	handle uint32
}

// Version returns the current semver FFI interface version.
func Version() string {
	v := pactffi_version()

	return goString(v)
}

var loggingInitialised string

// Init initialises the library.
func Init(logLevel string) {
	log.Println("[DEBUG] initialising native interface")
	logLevel = strings.ToUpper(logLevel)

	if loggingInitialised != "" {
		log.Printf("log level ('%s') cannot be set to '%s' after initialisation\n", loggingInitialised, logLevel)
		return
	}

	l, ok := logLevelStringToInt[logLevel]
	if !ok {
		l = LOG_LEVEL_INFO
	}
	log.Printf("[DEBUG] initialised native log level to %s (%d)", logLevel, l)

	if os.Getenv("PACT_LOG_PATH") != "" {
		log.Println("[DEBUG] initialised native log to log to file:", os.Getenv("PACT_LOG_PATH"))
		err := logToFile(os.Getenv("PACT_LOG_PATH"), l)
		if err != nil {
			log.Println("[ERROR] failed to log to file:", err)
		}
		return
	}

	log.Println("[DEBUG] initialised native log to log to stdout")
	err := logToStdout(l)
	if err != nil {
		log.Println("[ERROR] failed to log to stdout:", err)
	}
}

// MockServer is the public interface for managing the HTTP mock server.
type MockServer struct {
	pact         *Pact
	messagePact  *MessagePact
	interactions []*Interaction
}

// NewHTTPPact creates a new HTTP mock server for a given consumer/provider.
func NewHTTPPact(consumer string, provider string) *MockServer {
	return &MockServer{pact: &Pact{handle: pactffi_new_pact(consumer, provider)}}
}

// Version returns the current semver FFI interface version.
func (m *MockServer) Version() string {
	return Version()
}

// WithSpecificationVersion sets the Pact specification version the mock
// server writes and verifies interactions against.
func (m *MockServer) WithSpecificationVersion(version specificationVersion) {
	pactffi_with_specification(m.pact.handle, int32(version))
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es. Does not write the pact or cleanup server.
func (m *MockServer) Verify(port int, dir string) (bool, []MismatchedRequest) {
	mismatches := m.MockServerMismatchedRequests(port)
	log.Println("[DEBUG] mock server mismatches:", len(mismatches))

	return len(mismatches) == 0, mismatches
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MockServer) MockServerMismatchedRequests(port int) []MismatchedRequest {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []MismatchedRequest

	mismatches := pactffi_mock_server_mismatches(int32(port))
	// This method can return a nil pointer, in which case, it
	// should be considered a failure (or at least, an issue)
	// converting it to a string might also do nasty things here!
	if mismatches == nil {
		log.Println("[WARN] received a null pointer from the native interface, returning empty list of mismatches")
		return []MismatchedRequest{}
	}

	err := json.Unmarshal([]byte(goString(mismatches)), &res)
	if err != nil {
		log.Println("[ERROR] failed to unmarshal mismatches response, returning empty list of mismatches")
		return []MismatchedRequest{}
	}
	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MockServer) CleanupMockServer(port int) bool {
	if len(m.interactions) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := pactffi_cleanup_mock_server(int32(port))

	return res
}

// WritePactFile writes the Pact to file.
// TODO: expose overwrite.
func (m *MockServer) WritePactFile(port int, dir string) error {
	log.Println("[DEBUG] writing pact file for mock server on port:", port, ", dir:", dir)

	// overwritePact := 0
	// if overwrite {
	// 	overwritePact = 1
	// }

	// res := int(pactffi_write_pact_file(int32(port), dir, int32(overwritePact)))
	res := int(pactffi_write_pact_file(int32(port), dir, false))

	// | Error | Description |
	// |-------|-------------|
	// | 1 | A general panic was caught |
	// | 2 | The pact file was not able to be written |
	// | 3 | A mock server with the provided port was not found |
	switch res {
	case 0:
		return nil
	case 1:
		return ErrMockServerPanic
	case 2:
		return ErrUnableToWritePactFile
	case 3:
		return ErrMockServerNotfound
	default:
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// GetTLSConfig returns a tls.Config compatible with the TLS
// mock server.
func GetTLSConfig() *tls.Config {
	cert := pactffi_get_tls_ca_certificate()
	defer libRustFree(cert)

	goCert := goString(cert)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(goCert))

	return &tls.Config{
		RootCAs: certPool,
	}
}

func libRustFree(str unsafe.Pointer) {
	pactffi_free_string(str)
}

// Start starts up the mock HTTP server on the given address:port and TLS config
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MockServer) Start(address string, tls bool) (int, error) {
	if len(m.interactions) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address)

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return 0, ErrInvalidAddress
	}
	requestedPort, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, ErrInvalidAddress
	}
	var transport string
	if tls {
		transport = "https"
	} else {
		transport = "http"
	}

	msPort := int(pactffi_create_mock_server_for_transport(m.pact.handle, host, uint16(requestedPort), transport, nil))

	// | Error | Description |
	// |-------|-------------|
	// | -1 | A null pointer was received |
	// | -2 | The pact JSON could not be parsed |
	// | -3 | The mock server could not be started |
	// | -4 | The method panicked |
	// | -5 | The address is not valid |
	// | -6 | Could not create the TLS configuration with the self-signed certificate |
	switch msPort {
	case -1:
		return 0, ErrInvalidMockServerConfig
	case -2:
		return 0, ErrInvalidMockServerConfig
	case -3:
		return 0, ErrMockServerUnableToStart
	case -4:
		return 0, ErrMockServerPanic
	case -5:
		return 0, ErrInvalidAddress
	default:
		if msPort > 0 {
			log.Println("[DEBUG] mock server running on port:", msPort)
			return msPort, nil
		}
		return msPort, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", msPort)
	}
}

// StartTransport starts up a mock server on the given address:port for the given transport
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MockServer) StartTransport(transport string, address string, port int, config map[string][]any) (int, error) {
	if len(m.interactions) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address, port)

	configJSON := stringFromInterface(config)

	p := pactffi_create_mock_server_for_transport(m.pact.handle, address, uint16(port), transport, cstr(configJSON))

	// | Error | Description
	// |-------|-------------
	// | -1	   | An invalid handle was received. Handles should be created with pactffi_new_pact
	// | -2	   | transport_config is not valid JSON
	// | -3	   | The mock server could not be started
	// | -4	   | The method panicked
	// | -5	   | The address is not valid
	msPort := int(p)
	switch msPort {
	case -1:
		return 0, ErrInvalidMockServerConfig
	case -2:
		return 0, ErrInvalidMockServerConfig
	case -3:
		return 0, ErrMockServerUnableToStart
	case -4:
		return 0, ErrMockServerPanic
	case -5:
		return 0, ErrInvalidAddress
	default:
		if msPort > 0 {
			log.Println("[DEBUG] mock server running on port:", msPort)
			waitForTransport(address, msPort)
			return msPort, nil
		}
		return msPort, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", msPort)
	}
}

// WithMetadata sets additional metadata on the Pact file, grouped under
// namespace. Common uses are to add client library details such as the
// name and version.
func (m *MockServer) WithMetadata(namespace, k, v string) *MockServer {
	pactffi_with_pact_metadata(m.pact.handle, namespace, k, v)

	return m
}

// UsingPlugin loads a pact_ffi plugin by name and version so subsequent
// interactions on this pact can use plugin-provided matchers and content
// types (e.g. protobuf, gRPC).
func (m *MockServer) UsingPlugin(pluginName string, pluginVersion string) error {
	r := pactffi_using_plugin(m.pact.handle, pluginName, pluginVersion)

	// 1 - A general panic was caught.
	// 2 - Failed to load the plugin.
	// 3 - Pact Handle is not valid.
	res := int(r)
	switch res {
	case 1:
		return ErrPluginGenericPanic
	case 2:
		return ErrPluginFailed
	case 3:
		return ErrHandleNotFound
	default:
		if res != 0 {
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code", res)
		}
	}

	return nil
}

// CleanupPlugins releases the plugins loaded on this pact via UsingPlugin.
func (m *MockServer) CleanupPlugins() {
	pactffi_cleanup_plugins(m.pact.handle)
}

// NewInteraction initialises a new interaction for the current contract.
func (m *MockServer) NewInteraction(description string) *Interaction {
	i := &Interaction{
		handle: pactffi_new_interaction(m.pact.handle, description),
	}
	m.interactions = append(m.interactions, i)

	return i
}

// WithPluginInteractionContents sets the request or response contents of
// this interaction from a plugin-provided contentType (e.g. a protobuf
// message type), delegating the encoding to the loaded plugin.
func (i *Interaction) WithPluginInteractionContents(part interactionPart, contentType string, contents string) error {
	r := pactffi_interaction_contents(i.handle, int32(part), contentType, contents)

	// 1 - A general panic was caught.
	// 2 - The mock server has already been started.
	// 3 - The interaction handle is invalid.
	// 4 - The content type is not valid.
	// 5 - The contents JSON is not valid JSON.
	// 6 - The plugin returned an error.
	res := int(r)
	switch res {
	case 1:
		return ErrPluginGenericPanic
	case 2:
		return ErrPluginMockServerStarted
	case 3:
		return ErrPluginInteractionHandleInvalid
	case 4:
		return ErrPluginInvalidContentType
	case 5:
		return ErrPluginInvalidJSON
	case 6:
		return ErrPluginSpecificError
	default:
		if res != 0 {
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code", res)
		}
	}

	return nil
}

// UponReceiving sets the description of the request that triggers this
// interaction.
func (i *Interaction) UponReceiving(description string) *Interaction {
	pactffi_upon_receiving(i.handle, description)

	return i
}

// Given adds a provider state that must hold for this interaction.
func (i *Interaction) Given(state string) *Interaction {
	interactionGiven(i.handle, state)

	return i
}

// GivenWithParameter adds a provider state, parameterised by params, that
// must hold for this interaction.
func (i *Interaction) GivenWithParameter(state string, params map[string]any) *Interaction {
	interactionGivenWithParams(i.handle, state, params)

	return i
}

// WithRequest sets the request method and path (or path matcher) that
// this interaction expects.
func (i *Interaction) WithRequest(method string, pathOrMatcher any) *Interaction {
	path := stringFromInterface(pathOrMatcher)

	pactffi_with_request(i.handle, method, path)

	return i
}

// WithRequestHeaders sets the request headers this interaction expects,
// keyed by header name, with each value optionally a matcher.
func (i *Interaction) WithRequestHeaders(valueOrMatcher map[string][]any) *Interaction {
	return i.withHeaders(INTERACTION_PART_REQUEST, valueOrMatcher)
}

// WithResponseHeaders sets the response headers this interaction returns,
// keyed by header name, with each value optionally a matcher.
func (i *Interaction) WithResponseHeaders(valueOrMatcher map[string][]any) *Interaction {
	return i.withHeaders(INTERACTION_PART_RESPONSE, valueOrMatcher)
}

// WithQuery sets the request query parameters this interaction expects,
// keyed by parameter name, with each value optionally a matcher.
func (i *Interaction) WithQuery(valueOrMatcher map[string][]any) *Interaction {
	for k, values := range valueOrMatcher {
		for idx, v := range values {
			value := stringFromInterface(v)

			pactffi_with_query_parameter_v2(i.handle, k, uintptr(idx), value)
		}
	}

	return i
}

// WithJSONRequestBody sets the request body, JSON-encoding body (which may
// contain matchers).
func (i *Interaction) WithJSONRequestBody(body any) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_REQUEST)
}

// WithJSONResponseBody sets the response body, JSON-encoding body (which
// may contain matchers).
func (i *Interaction) WithJSONResponseBody(body any) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_RESPONSE)
}

// WithRequestBody sets the raw request body and its content type.
func (i *Interaction) WithRequestBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 0)
}

// WithResponseBody sets the raw response body and its content type.
func (i *Interaction) WithResponseBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 1)
}

// WithBinaryRequestBody sets the request body to the given bytes, with an
// application/octet-stream content type.
func (i *Interaction) WithBinaryRequestBody(body []byte) *Interaction {
	return i.withBinaryBody("application/octet-stream", body, INTERACTION_PART_REQUEST)
}

// WithBinaryResponseBody sets the response body to the given bytes, with
// an application/octet-stream content type.
func (i *Interaction) WithBinaryResponseBody(body []byte) *Interaction {
	return i.withBinaryBody("application/octet-stream", body, INTERACTION_PART_RESPONSE)
}

// WithRequestMultipartFile sets the request body to a multipart file
// upload, reading filename under mimePartName.
func (i *Interaction) WithRequestMultipartFile(contentType string, filename string, mimePartName string) *Interaction {
	return i.withMultipartFile(contentType, filename, mimePartName, INTERACTION_PART_REQUEST)
}

// WithResponseMultipartFile sets the response body to a multipart file
// upload, reading filename under mimePartName.
func (i *Interaction) WithResponseMultipartFile(contentType string, filename string, mimePartName string) *Interaction {
	return i.withMultipartFile(contentType, filename, mimePartName, INTERACTION_PART_RESPONSE)
}

// WithStatus sets the expected HTTP response status.
func (i *Interaction) WithStatus(status int) *Interaction {
	pactffi_response_status(i.handle, uint16(status))

	return i
}

// WithReference records an external reference (e.g. a ticket or pull request)
// against the interaction. References are stored under comments.references[group][name]
// in the Pact file. This is a V4-only feature.
func (i *Interaction) WithReference(group, name, value string) *Interaction {
	pactffi_add_interaction_reference(i.handle, group, name, value)

	return i
}

func (i *Interaction) withHeaders(part interactionPart, valueOrMatcher map[string][]any) *Interaction {
	for k, v := range valueOrMatcher {
		for idx, header := range v {
			value := stringFromInterface(header)

			pactffi_with_header_v2(i.handle, int32(part), k, uintptr(idx), value)
		}
	}

	return i
}

func (i *Interaction) withJSONBody(body any, part interactionPart) *Interaction {
	jsonBody := stringFromInterface(body)

	pactffi_with_body(i.handle, int32(part), "application/json", jsonBody)

	return i
}

func (i *Interaction) withBody(contentType string, body []byte, part interactionPart) *Interaction {
	pactffi_with_body(i.handle, int32(part), contentType, string(body))

	return i
}

func (i *Interaction) withBinaryBody(contentType string, body []byte, part interactionPart) *Interaction {
	pactffi_with_binary_file(i.handle, int32(part), contentType, unsafe.Pointer(&body[0]), uintptr(len(body)))

	return i
}

func (i *Interaction) withMultipartFile(contentType string, filename string, mimePartName string, part interactionPart) *Interaction {
	pactffi_with_multipart_file(i.handle, int32(part), contentType, filename, mimePartName)

	return i
}

type stringLike interface {
	String() string
}

func stringFromInterface(obj any) string {
	switch t := obj.(type) {
	case string:
		return t
	default:
		bytes, err := json.Marshal(obj)
		if err != nil {
			panic(fmt.Sprintln("unable to marshal body to JSON:", err))
		}
		return quotedString(string(bytes))
	}
}

// This fixes a quirk where certain "matchers" (e.g. matchers.S/String) are
// really just strings. However, whene we JSON encode them they get wrapped in quotes
// and the rust core sees them as plain strings, requiring then the quotes to be matched.
func quotedString(s string) string {
	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Experimental logging options
// func LogMessage(pkg, level, message string) {
// 	cPkg := C.CString(pkg)
// 	defer free(cPkg)

// 	cLevel := C.CString(level)
// 	defer free(cLevel)

// 	cMessage := C.CString(message)
// 	defer free(cMessage)

// 	res := pactffi_log_message(cPkg, cLevel, cMessage)
// 	log.Println("[DEBUG] log_to_buffer res", res)
// }

// func logToBuffer(level logLevel) error {
// 	res := pactffi_log_to_buffer(int32(level))
// 	log.Println("[DEBUG] log_to_buffer res", res)

// 	return logResultToError(int(res))
// }

func logToStdout(level logLevel) error {
	res := pactffi_log_to_stdout(int32(level))
	log.Println("[DEBUG] log_to_stdout res", res)

	return logResultToError(int(res))
}

func logToFile(file string, level logLevel) error {
	res := pactffi_log_to_file(file, int32(level))
	log.Println("[DEBUG] log_to_file res", res)

	return logResultToError(int(res))
}

// func getLogBuffer() string {
// 	buf := pactffi_fetch_log_buffer()
// 	defer free(buf)

// 	return goString(buf)
// }

func logResultToError(res int) error {
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCantSetLogger
	case -2:
		return ErrNoLogger
	case -3:
		return ErrSpecifierNotUtf8
	case -4:
		return ErrUnknownSinkType
	case -5:
		return ErrMissingFilePath
	case -6:
		return ErrCantOpenSinkToFile
	case -7:
		return ErrCantConstructSink
	default:
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// Errors.
var (
	// ErrHandleNotFound indicates the underlying handle was not found, and a logic error in the framework.
	ErrHandleNotFound = errors.New("unable to find the native interface handle (this indicates a defect in the framework)")

	// ErrMockServerPanic indicates a panic occurred when invoking the remote Mock Server.
	ErrMockServerPanic = errors.New("a general panic occurred when starting/invoking mock service (this indicates a defect in the framework)")

	// ErrUnableToWritePactFile indicates an error when writing the pact file to disk.
	ErrUnableToWritePactFile = errors.New("unable to write to file")

	// ErrMockServerNotfound indicates the Mock Server could not be found.
	ErrMockServerNotfound = errors.New("unable to find mock server with the given port")

	// ErrInvalidMockServerConfig indicates an issue configuring the mock server.
	ErrInvalidMockServerConfig = errors.New("configuration for the mock server was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	// ErrInvalidPact indicates the pact file provided to the mock server was not a valid pact file.
	ErrInvalidPact = errors.New("pact given to mock server is invalid")

	// ErrMockServerUnableToStart means the mock server could not be started in the rust library.
	ErrMockServerUnableToStart = errors.New("unable to start the mock server")

	// ErrInvalidAddress means the address provided to the mock server was invalid and could not be understood.
	ErrInvalidAddress = errors.New("invalid address provided to the mock server")

	// ErrMockServerTLSConfiguration indicates a TLS mock server could not be started
	// and is likely a framework level problem.
	ErrMockServerTLSConfiguration = errors.New("a tls mock server could not be started (this is likely a defect in the framework)")

	// ErrNoInteractions indicates no Interactions have been registered to a mock server, and cannot be started/stopped until at least one is added.
	ErrNoInteractions = errors.New("no interactions have been registered for the mock server")

	// ErrPluginFailed indicates the plugin could not be started.
	ErrPluginFailed = errors.New("the plugin could not be started")
)

// Log Errors.
var (
	ErrCantSetLogger      = errors.New("can't set logger (applying the logger failed, perhaps because one is applied already)")
	ErrNoLogger           = errors.New("no logger has been initialized (call `logger_init` before any other log function)")
	ErrSpecifierNotUtf8   = errors.New("the sink specifier was not UTF-8 encoded")
	ErrUnknownSinkType    = errors.New(`the sink type specified is not a known type (known types: "buffer", "stdout", "stderr", or "file /some/path")`)
	ErrMissingFilePath    = errors.New("no file path was specified in a file-type sink specification")
	ErrCantOpenSinkToFile = errors.New("opening a sink to the specified file path failed (check permissions)")
	ErrCantConstructSink  = errors.New("can't construct the log sink")
)
