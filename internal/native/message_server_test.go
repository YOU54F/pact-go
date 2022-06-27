package native

import (
	"bytes"
	"compress/gzip"
	context "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	l "log"
	"os"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/log"

	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestHandleBasedMessageTestsWithString(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	s := NewMessageServer("test-message-consumer", "test-message-provider")

	m := s.NewMessage().
		Given("some state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithContents("text/plain", []byte("some string"))

	body := m.ReifyMessage()

	var res jsonMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	assert.Equal(t, res.Description, "some message")
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithJSON(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	s := NewMessageServer("test-message-consumer", "test-message-provider")

	m := s.NewMessage().
		Given("some state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithJSONContents(map[string]string{
			"some": "json",
		})

	body := m.ReifyMessage()
	l.Println(body) // TODO: JSON is not stringified - probably should be?

	var res jsonMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	assert.Equal(t, res.Description, "some message")
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithBinary(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	s := NewMessageServer("test-binarymessage-consumer", "test-binarymessage-provider").
		WithMetadata("some-namespace", "the-key", "the-value")

	// generate some binary data
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	encodedMessage := "A long time ago in a galaxy far, far away..."
	_, err = zw.Write([]byte(encodedMessage))
	assert.NoError(t, err)

	err = zw.Close()
	assert.NoError(t, err)

	m := s.NewMessage().
		Given("some binary state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some binary message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithBinaryContents(buf.Bytes())

	body := m.ReifyMessage()

	// Check the reified message is good

	var res binaryMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	// Extract binary payload, base 64 decode it, unzip it
	data, err := base64.RawStdEncoding.DecodeString(res.Contents)
	assert.NoError(t, err)
	r, err := gzip.NewReader(bytes.NewReader(data))
	assert.NoError(t, err)
	result, _ := ioutil.ReadAll(r)

	assert.Equal(t, encodedMessage, string(result))
	assert.Equal(t, "some binary message", res.Description)
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message...

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

type binaryMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       string                   `json:"contents"` // base 64 encoded
	// Contents       []byte                   `json:"contents"`
}
type jsonMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       interface{}              `json:"contents"`
}

func TestGrpcPluginInteraction(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	log.InitLogging()
	log.SetLogLevel("TRACE")

	m := NewMessageServer("test-message-consumer", "test-message-provider")
	// m := NewPact("test-grpc-consumer", "test-plugin-provider")

	// Protobuf plugin test
	m.UsingPlugin("protobuf", "0.1.7")
	// m.WithSpecificationVersion(SPECIFICATION_VERSION_V4)

	i := m.NewSyncMessageInteraction("grpc interaction")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/plugin.proto", dir)

	grpcInteraction := `{
			"pact:proto": "` + path + `",
			"pact:proto-service": "PactPlugin/InitPlugin",
			"pact:content-type": "application/protobuf",
			"request": {
				"implementation": "notEmpty('pact-go-driver')",
				"version": "matching(semver, '0.0.0')"	
			},
			"response": {
				"catalogue": [
					{
						"type": "INTERACTION",
						"key": "test"
					}
				]
			}
		}`

	i.
		Given("plugin state").
		// For gRPC interactions we prpvide the config once for both the request and response parts
		WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)

	// Start the gRPC mock server
	port, err := m.StartTransport("grpc", "127.0.0.1", 0, make(map[string][]interface{}))
	assert.NoError(t, err)
	defer m.CleanupMockServer(port)

	// Now we can make a normal gRPC request
	initPluginRequest := &InitPluginRequest{
		Implementation: "pact-go-test",
		Version:        "1.0.0",
	}

	// Need to make a gRPC call here
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewPactPluginClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.InitPlugin(ctx, initPluginRequest)
	if err != nil {
		l.Fatalf("could not initialise the plugin: %v", err)
	}
	l.Printf("InitPluginResponse: %v", r)

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 0 {
		assert.Len(t, mismatches, 0)
		t.Log(mismatches)
	}

	err = m.WritePactFileForServer(port, tmpPactFolder, true)
	assert.NoError(t, err)
}

// TODO:

// In the test, you use `NewSyncMessageInteraction` but don't set a response type on the FFI, just the plugin. This is a little awkward
// Does this imply that a sync can be async?
// WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)

// // How to add metadata, binary contents ?

// // -> https://docs.pact.io/slack/libpact_ffi-users.html#1646812690.612989

// Is the mental model to Ignore the V3 message pact stuff now (MessagePactHandle and MessageHandle) and just assume `PactHandle` and `InteractionHandle` are the main thinngs?
// If so, is it possible to get an updated view of which functions are essentially deprecated? I'm finding it very hard to know which methods to call.

// For example,

// MessagePactHandle                     PactHandle
// pactffi_with_message_pact_metadata --> pactffi_with_pact_metadata

// MessageHandle                      --> InteractionHandle
// pactffi_with_message_pact_metadata --> pactffi_with_binary_file This would be confusing though

// Example of methods on Interaction that don't make sense for all types. This means all client libraries need to know the cases in which it can/can't be used
// pactffi_response_status, pactffi_with_query_parameter, pactffi_with_request, pactffi_with_binary_file

// Example of methods on MessageHandle that don't have equivalent on InteractionHandle
// pactffi_message_expects_to_receive, pactffi_message_with_contents, pactffi_message_reify