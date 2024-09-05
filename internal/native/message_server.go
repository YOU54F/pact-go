package native

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"unsafe"
)

// MessagePact is a Go representation of the message-based PactHandle struct.
type MessagePact struct {
	handle uint16
}

type messageType int

// Whether a Message was created as an asynchronous or synchronous
// (request/response) message interaction.
const (
	MESSAGE_TYPE_ASYNC messageType = iota
	MESSAGE_TYPE_SYNC
)

// Message is a single message interaction on a MessageServer, either
// asynchronous (one body) or synchronous (a request/response pair).
type Message struct {
	handle      uint32
	messageType messageType
	pact        *MessagePact
	index       int
	server      *MessageServer
}

// MessageServer is the public interface for managing the message based interface.
type MessageServer struct {
	messagePact *MessagePact
	messages    []*Message
}

// NewMessageServer creates a new message-based Pact for a given
// consumer/provider.
func NewMessageServer(consumer string, provider string) *MessageServer {
	return &MessageServer{messagePact: &MessagePact{handle: pactffi_new_message_pact(consumer, provider)}}
}

// WithMetadata sets additional metadata on the Pact file, grouped under
// namespace. Common uses are to add client library details such as the
// name and version.
func (m *MessageServer) WithMetadata(namespace, k, v string) *MessageServer {
	pactffi_with_message_pact_metadata(m.messagePact.handle, namespace, k, v)

	return m
}

// NewMessage initialises a new message for the current contract.
//
// Deprecated: use NewAsyncMessageInteraction instead.
func (m *MessageServer) NewMessage() *Message {
	// Alias
	return m.NewAsyncMessageInteraction("")
}

// NewSyncMessageInteraction initialises a new synchronous message interaction for the current contract.
func (m *MessageServer) NewSyncMessageInteraction(description string) *Message {
	i := &Message{
		handle:      pactffi_new_sync_message_interaction(m.messagePact.handle, description),
		messageType: MESSAGE_TYPE_SYNC,
		pact:        m.messagePact,
		index:       len(m.messages),
		server:      m,
	}
	m.messages = append(m.messages, i)

	return i
}

// NewAsyncMessageInteraction initialises a new asynchronous message interaction for the current contract.
func (m *MessageServer) NewAsyncMessageInteraction(description string) *Message {
	i := &Message{
		handle:      pactffi_new_message_interaction(m.messagePact.handle, description),
		messageType: MESSAGE_TYPE_ASYNC,
		pact:        m.messagePact,
		index:       len(m.messages),
		server:      m,
	}
	m.messages = append(m.messages, i)

	return i
}

// WithSpecificationVersion sets the Pact specification version this
// message pact is written and verified against.
func (m *MessageServer) WithSpecificationVersion(version specificationVersion) {
	pactffi_with_specification(m.messagePact.handle, int32(version))
}

// Given adds a provider state that must hold for this message.
func (m *Message) Given(state string) *Message {
	interactionGiven(m.handle, state)

	return m
}

// GivenWithParameter adds a provider state, parameterised by params, that
// must hold for this message.
func (m *Message) GivenWithParameter(state string, params map[string]any) *Message {
	if len(params) == 0 {
		interactionGiven(m.handle, state)
	} else {
		interactionGivenWithParams(m.handle, state, params)
	}

	return m
}

// ExpectsToReceive sets the description of this message interaction.
func (m *Message) ExpectsToReceive(description string) *Message {
	pactffi_message_expects_to_receive(m.handle, description)

	return m
}

// WithMetadata sets metadata key/value pairs to expect alongside this
// message's contents.
func (m *Message) WithMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		// TODO: check if matching rules allowed here
		// value := stringFromInterface(v)
		// fmt.Printf("withheaders, sending: %+v \n\n", value)
		// cValue := C.CString(value)

		pactffi_message_with_metadata(m.handle, k, v)
	}

	return m
}

// WithRequestMetadata sets metadata key/value pairs to expect on the
// request half of this synchronous message.
func (m *Message) WithRequestMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		pactffi_with_metadata(m.handle, k, v, 0)
	}

	return m
}

// WithResponseMetadata sets metadata key/value pairs to expect on the
// response half of this synchronous message.
func (m *Message) WithResponseMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		pactffi_with_metadata(m.handle, k, v, 1)
	}

	return m
}

// WithRequestBinaryContents sets the request contents of this message to
// body, with an application/octet-stream content type.
func (m *Message) WithRequestBinaryContents(body []byte) *Message {
	// TODO: handle response
	res := pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_REQUEST), "application/octet-stream", unsafe.Pointer(&body[0]), uintptr(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}

// WithRequestBinaryContentType sets the request contents of this message
// to body, with the given content type.
func (m *Message) WithRequestBinaryContentType(contentType string, body []byte) *Message {
	// TODO: handle response
	res := pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_REQUEST), contentType, unsafe.Pointer(&body[0]), uintptr(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}

// WithRequestJSONContents JSON-encodes body (which may contain matchers)
// as the request contents of this message.
func (m *Message) WithRequestJSONContents(body any) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_REQUEST, "application/json", []byte(value))
}

// WithResponseBinaryContents sets the response contents of this
// synchronous message to body, with an application/octet-stream content
// type.
func (m *Message) WithResponseBinaryContents(body []byte) *Message {
	// TODO: handle response
	pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_RESPONSE), "application/octet-stream", unsafe.Pointer(&body[0]), uintptr(len(body)))

	return m
}

// WithResponseJSONContents JSON-encodes body (which may contain matchers)
// as the response contents of this synchronous message.
func (m *Message) WithResponseJSONContents(body any) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_RESPONSE, "application/json", []byte(value))
}

// WithContents sets the request or response contents of this message to
// body with the given content type. Note that string values here must be
// NUL terminated.
func (m *Message) WithContents(part interactionPart, contentType string, body []byte) *Message {
	res := pactffi_with_body(m.handle, int32(part), contentType, string(body))
	log.Println("[DEBUG] response from pactffi_interaction_contents", res)

	return m
}

// TODO: migrate plugin code to shared struct/code?

// UsingPlugin loads a pact_ffi plugin by name and version so subsequent
// messages on this pact can use plugin-provided matchers and content
// types (e.g. protobuf, gRPC).
func (m *MessageServer) UsingPlugin(pluginName string, pluginVersion string) error {
	r := pactffi_using_plugin(m.messagePact.handle, pluginName, pluginVersion)

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

// WithPluginInteractionContents sets the request or response contents of
// this message from a plugin-provided contentType (e.g. a protobuf
// message type), delegating the encoding to the loaded plugin.
func (m *Message) WithPluginInteractionContents(part interactionPart, contentType string, contents string) error {
	r := pactffi_interaction_contents(m.handle, int32(part), contentType, contents)

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

// GetMessageRequestContents retrieves the binary contents of the request
// for a given message; any matchers are stripped away if given. If the
// contents are from a plugin, the byte[] representation of the parsed
// plugin data is returned, again with any matchers etc. removed.
func (m *Message) GetMessageRequestContents() ([]byte, error) {
	log.Println("[DEBUG] GetMessageRequestContents")
	if m.messageType == MESSAGE_TYPE_ASYNC {
		return m.getAsyncMessageRequestContents()
	}
	return m.getSyncMessageRequestContents()
}

// GetMessageResponseContents retreives the binary contents of the response for a given message
// any matchers are stripped away if given
// if the contents is from a plugin, the byte[] representation of the parsed
// plugin data is returned, again, with any matchers etc. removed.
func (m *Message) GetMessageResponseContents() ([][]byte, error) {
	responses := make([][]byte, len(m.server.messages))
	if m.messageType == MESSAGE_TYPE_ASYNC {
		return nil, errors.New("invalid request: asynchronous messages do not have response")
	}
	iter := pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}
	defer pactffi_pact_sync_message_iter_delete(iter)

	for i := range len(m.server.messages) {
		message := pactffi_pact_sync_message_iter_next(iter)

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		// Get Response body
		contentsLen := pactffi_sync_message_get_response_contents_length(message, uintptr(i))
		if contentsLen != 0 {
			data := pactffi_sync_message_get_response_contents_bin(message, uintptr(i))
			if data == nil {
				return nil, errors.New("retrieved an empty pointer to the message contents")
			}
			bytes := goBytes(data, int(contentsLen))
			responses[i] = bytes
		}
	}

	return responses, nil
}

// StartTransport starts up a mock server on the given address:port for the given transport
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MessageServer) StartTransport(transport string, address string, port int, config map[string][]any) (int, error) {
	if len(m.messages) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address, port)

	configJSON := stringFromInterface(config)

	p := pactffi_create_mock_server_for_transport(m.messagePact.handle, address, uint16(port), transport, cstr(configJSON))

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

// CleanupPlugins releases the plugins loaded on this pact via UsingPlugin.
func (m *MessageServer) CleanupPlugins() {
	pactffi_cleanup_plugins(m.messagePact.handle)
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MessageServer) CleanupMockServer(port int) bool {
	if len(m.messages) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := pactffi_cleanup_mock_server(int32(port))

	return res
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MessageServer) MockServerMismatchedRequests(port int) []MismatchedRequest {
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

// MockServerMatched reports whether every interaction registered against
// the mock server on port was matched by an actual request.
func (m *MessageServer) MockServerMatched(port int) bool {
	log.Println("[DEBUG] mock server determining mismatches:", port)

	res := pactffi_mock_server_matched(int32(port))

	// TODO: why this number is so big and not a bool? Type def wrong? Port value wrong?
	// log.Println("MATCHED RES?")
	// log.Println(int(res))

	return res
}

// WritePactFile writes the Pact to file.
func (m *MessageServer) WritePactFile(dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)

	overwritePact := overwrite

	res := int(pactffi_write_message_pact_file(m.messagePact.handle, dir, overwritePact))

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The pact file was not able to be written |
	/// | 2 | The message pact for the given handle was not found |
	switch res {
	case 0:
		return nil
	case 1:
		return ErrUnableToWritePactFile
	case 2:
		return ErrHandleNotFound
	default:
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// WritePactFileForServer writes the Pact for the mock server running on
// port to dir, optionally overwriting an existing file.
func (m *MessageServer) WritePactFileForServer(port int, dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)

	overwritePact := overwrite

	res := int(pactffi_write_pact_file(int32(port), dir, overwritePact))

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The pact file was not able to be written |
	/// | 2 | The message pact for the given handle was not found |
	switch res {
	case 0:
		return nil
	case 1:
		return ErrMockServerPanic
	case 2:
		return ErrUnableToWritePactFile
	case 3:
		return ErrHandleNotFound
	default:
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// WithReference records an external reference (e.g. a ticket or pull request)
// against the interaction. References are stored under comments.references[group][name]
// in the Pact file. This is a V4-only feature.
func (m *Message) WithReference(group, name, value string) *Message {
	pactffi_add_interaction_reference(m.handle, group, name, value)

	return m
}

// getAsyncMessageRequestContents is the MESSAGE_TYPE_ASYNC branch of
// GetMessageRequestContents, split out to keep both branches readable.
func (m *Message) getAsyncMessageRequestContents() ([]byte, error) {
	iter := pactffi_pact_handle_get_message_iter(m.pact.handle)
	log.Println("[DEBUG] pactffi_pact_handle_get_message_iter")
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}
	defer pactffi_pact_message_iter_delete(iter)
	log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - OK")

	///////
	// TODO: some debugging in here to see what's exploding.......
	///////

	log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - len", len(m.server.messages))

	for i := range len(m.server.messages) {
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - index", i)
		message := pactffi_pact_message_iter_next(iter)
		log.Println("[DEBUG] pactffi_pact_message_iter_next - message", message)

		if i != m.index {
			continue
		}
		log.Println("[DEBUG] pactffi_pact_message_iter_next - index match", message)

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		contentsLen := pactffi_message_get_contents_length(message)
		log.Println("[DEBUG] pactffi_message_get_contents_length - len", contentsLen)
		if contentsLen == 0 {
			// You can have empty bodies
			log.Println("[DEBUG] message body is empty")
			return nil, nil
		}
		data := pactffi_message_get_contents_bin(message)
		log.Println("[DEBUG] pactffi_message_get_contents_bin - data", data)
		if data == nil {
			// You can have empty bodies
			log.Println("[DEBUG] message binary contents are empty")
			return nil, nil
		}
		bytes := goBytes(data, int(contentsLen))

		return bytes, nil
	}

	return nil, errors.New("unable to find the message")
}

// getSyncMessageRequestContents is the synchronous-message branch of
// GetMessageRequestContents, split out to keep both branches readable.
func (m *Message) getSyncMessageRequestContents() ([]byte, error) {
	iter := pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}
	defer pactffi_pact_sync_message_iter_delete(iter)

	for i := range len(m.server.messages) {
		message := pactffi_pact_sync_message_iter_next(iter)

		if i != m.index {
			continue
		}

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		contentsLen := pactffi_sync_message_get_request_contents_length(message)
		if contentsLen == 0 {
			log.Println("[DEBUG] message body is empty")
			return nil, nil
		}
		data := pactffi_sync_message_get_request_contents_bin(message)
		if data == nil {
			log.Println("[DEBUG] message binary contents are empty")
			return nil, nil
		}
		bytes := goBytes(data, int(contentsLen))

		return bytes, nil
	}

	return nil, errors.New("unable to find the message")
}
