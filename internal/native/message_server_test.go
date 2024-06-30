package native

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleBasedMessageTestsWithString(t *testing.T) {
	tmpPactFolder, err := os.MkdirTemp("", "pact-go")
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
		WithContents(INTERACTION_PART_REQUEST, "text/plain", []byte("some string"))

	body, err := m.GetMessageRequestContents()

	assert.NoError(t, err)
	assert.Equal(t, "some string", string(body))

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithJSON(t *testing.T) {
	tmpPactFolder, err := os.MkdirTemp("", "pact-go")
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
		WithRequestJSONContents(map[string]string{
			"some": "json",
		})

	body, err := m.GetMessageRequestContents()
	assert.NoError(t, err)

	var res struct {
		Some string `json:"some"`
	}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)
	assert.Equal(t, "json", res.Some)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithBinary(t *testing.T) {
	tmpPactFolder, err := os.MkdirTemp("", "pact-go")
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
		WithRequestBinaryContentType("application/gzip", buf.Bytes())

	body, err := m.GetMessageRequestContents()
	assert.NoError(t, err)

	// Extract binary payload, base 64 decode it, unzip it
	r, err := gzip.NewReader(bytes.NewReader(body))
	assert.NoError(t, err)
	result, _ := io.ReadAll(r)

	assert.Equal(t, encodedMessage, string(result))

	// This is where you would invoke the real function with the message...

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestGetAsyncMessageContentsAsBytes(t *testing.T) {
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
		WithRequestJSONContents(map[string]string{
			"some": "json",
		})

	bytes, err := m.GetMessageRequestContents()
	assert.NoError(t, err)
	assert.NotNil(t, bytes)

	// Should be able to convert back into the JSON structure
	var v struct {
		Some string `json:"some"`
	}
	err = json.Unmarshal(bytes, &v)
	assert.NoError(t, err)
	assert.Equal(t, "json", v.Some)
}

// func TestGetSyncMessageContentsAsBytes(t *testing.T) {
// 	s := NewMessageServer("test-message-consumer", "test-message-provider")

// 	m := s.NewSyncMessageInteraction("").
// 		Given("some state").
// 		GivenWithParameter("param", map[string]interface{}{
// 			"foo": "bar",
// 		}).
// 		ExpectsToReceive("some message").
// 		WithMetadata(map[string]string{
// 			"meta": "data",
// 		}).
// 		// WithResponseJSONContents(map[string]string{
// 		// 	"some": "request",
// 		// }).
// 		WithResponseJSONContents(map[string]string{
// 			"some": "response",
// 		})

// 	bytes, err := m.GetMessageResponseContents()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, bytes)

// 	// Should be able to convert back into the JSON structure
// 	var v struct {
// 		Some string `json:"some"`
// 	}
// 	err = json.Unmarshal(bytes[0], &v)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "response", v.Some)
// }

func TestGetSyncMessageContentsAsBytes_EmptyResponse(t *testing.T) {
	s := NewMessageServer("test-message-consumer", "test-message-provider")

	m := s.NewSyncMessageInteraction("").
		Given("some state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some message").
		WithMetadata(map[string]string{
			"meta": "data",
		})

	bytes, err := m.GetMessageResponseContents()
	assert.NoError(t, err)
	assert.NotNil(t, bytes)
	assert.Equal(t, 1, len(bytes))
	assert.Empty(t, bytes[0])
}
