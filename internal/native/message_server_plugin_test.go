//go:build plugin
// +build plugin

package native

import (
	context "context"
	"fmt"
	l "log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// func TestGetPluginSyncMessageContentsAsBytes(t *testing.T) {
// 	if runtime.GOOS != "windows" {
// 		m := NewMessageServer("test-message-consumer", "test-message-provider")

// 		// Protobuf plugin test
// 		err := m.UsingPlugin("protobuf", "0.3.15")
// 		defer m.CleanupPlugins()
// 		assert.NoError(t, err)

// 		i := m.NewSyncMessageInteraction("grpc interaction")

// 		dir, _ := os.Getwd()
// 		path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

// 		grpcInteraction := `{
// 			"pact:proto": "` + path + `",
// 			"pact:proto-service": "PactPlugin/InitPlugin",
// 			"pact:content-type": "application/protobuf",
// 			"request": {
// 				"implementation": "notEmpty('pact-go-driver')",
// 				"version": "matching(semver, '0.0.0')"
// 			},
// 			"response": {
// 				"catalogue": [
// 					{
// 						"type": "INTERACTION",
// 						"key": "test"
// 					}
// 				]
// 			}
// 		}`

// 		err = i.
// 			Given("plugin state").
// 			// For gRPC interactions we prpvide the config once for both the request and response parts
// 			WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)
// 		assert.NoError(t, err)

// 		bytes, err := i.GetMessageRequestContents()
// 		assert.NoError(t, err)
// 		assert.NotNil(t, bytes)

// 		// Should be able to convert request body back into a protobuf
// 		p := &InitPluginRequest{}
// 		err = proto.Unmarshal(bytes, p)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "0.0.0", p.Version)

// 		// Should be able to convert response into a protobuf
// 		response, err := i.GetMessageResponseContents()
// 		assert.NoError(t, err)
// 		assert.NotNil(t, bytes)
// 		r := &InitPluginResponse{}
// 		err = proto.Unmarshal(response[0], r)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "test", r.Catalogue[0].Key)
// 	}
// }

// func TestGetPluginSyncMessageContentsAsBytes_EmptyResponse(t *testing.T) {
// 	if runtime.GOOS != "windows" {
// 		m := NewMessageServer("test-message-consumer", "test-message-provider")

// 		// Protobuf plugin test
// 		err := m.UsingPlugin("protobuf", "0.3.15")
// 		defer m.CleanupPlugins()
// 		assert.NoError(t, err)

// 		i := m.NewSyncMessageInteraction("grpc interaction")

// 		dir, _ := os.Getwd()
// 		path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

// 		grpcInteraction := `{
// 			"pact:proto": "` + path + `",
// 			"pact:proto-service": "PactPlugin/InitPlugin",
// 			"pact:content-type": "application/protobuf",
// 			"request": {
// 				"implementation": "notEmpty('pact-go-driver')",
// 				"version": "matching(semver, '0.0.0')"
// 			}
// 		}`

// 		err = i.
// 			Given("plugin state").
// 			// For gRPC interactions we prpvide the config once for both the request and response parts
// 			WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)
// 		assert.NoError(t, err)

// 		bytes, err := i.GetMessageRequestContents()
// 		assert.NoError(t, err)
// 		assert.NotNil(t, bytes)

// 		// Should be able to convert request body back into a protobuf
// 		p := &InitPluginRequest{}
// 		err = proto.Unmarshal(bytes, p)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "0.0.0", p.Version)

// 		// Should be able to convert response into a protobuf
// 		response_bytes, err := i.GetMessageResponseContents()
// 		assert.NoError(t, err)
// 		assert.NotNil(t, response_bytes)
// 		assert.Equal(t, 1, len(response_bytes))
// 		assert.Empty(t, response_bytes[0])
// 	}
// }

func TestGetPluginAsyncMessageContentsAsBytes(t *testing.T) {
	if runtime.GOOS != "windows" {
		m := NewMessageServer("test-message-consumer", "test-message-provider")

		// Protobuf plugin test
		_ = m.UsingPlugin("protobuf", "0.3.15")
		defer m.CleanupPlugins()

		i := m.NewAsyncMessageInteraction("grpc interaction")

		dir, _ := os.Getwd()
		path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

		protobufInteraction := `{
			"pact:proto": "` + path + `",
			"pact:message-type": "InitPluginRequest",
			"pact:content-type": "application/protobuf",
			"implementation": "notEmpty('pact-go-driver')",
			"version": "matching(semver, '0.0.0')"
		}`

		err := i.
			Given("plugin state").
			// For gRPC interactions we prpvide the config once for both the request and response parts
			WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", protobufInteraction)
		assert.NoError(t, err)

		bytes, err := i.GetMessageRequestContents()
		assert.NoError(t, err)
		assert.NotNil(t, bytes)

		// Should be able to convert body back into a protobuf
		p := &InitPluginRequest{}
		err = proto.Unmarshal(bytes, p)
		assert.NoError(t, err)
		assert.Equal(t, "0.0.0", p.Version)
	}
}

func TestGrpcPluginInteraction(t *testing.T) {
	if runtime.GOOS != "windows" {
		tmpPactFolder, err := os.MkdirTemp("", "pact-go")
		assert.NoError(t, err)
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "TRACE"
		}
		_ = log.SetLogLevel(logutils.LogLevel(logLevel))

		m := NewMessageServer("test-message-consumer", "test-message-provider")

		// Protobuf plugin test
		_ = m.UsingPlugin("protobuf", "0.3.15")
		defer m.CleanupPlugins()

		i := m.NewSyncMessageInteraction("grpc interaction")

		dir, _ := os.Getwd()
		path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

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

		err = i.
			Given("plugin state").
			// For gRPC interactions we prpvide the config once for both the request and response parts
			WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)
		assert.NoError(t, err)

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
}

func TestGrpcPluginInteraction_ErrorResponse(t *testing.T) {
	if runtime.GOOS != "windows" {
		tmpPactFolder, err := os.MkdirTemp("", "pact-go")
		assert.NoError(t, err)
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "TRACE"
		}
		_ = log.SetLogLevel(logutils.LogLevel(logLevel))

		m := NewMessageServer("test-message-consumer", "test-message-provider")

		// Protobuf plugin test
		_ = m.UsingPlugin("protobuf", "0.3.15")

		i := m.NewSyncMessageInteraction("grpc interaction")

		dir, _ := os.Getwd()
		path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

		grpcInteraction := `{
			"pact:proto": "` + path + `",
			"pact:proto-service": "PactPlugin/InitPlugin",
			"pact:content-type": "application/protobuf",
			"request": {
				"implementation": "notEmpty('pact-go-driver')",
				"version": "matching(semver, '0.0.0')"
			},
			"responseMetadata": {
				"grpc-status": "NOT_FOUND",
				"grpc-message": "matching(type, 'not found')"
			}
		}`

		err = i.
			Given("plugin state").
			// For gRPC interactions we prpvide the config once for both the request and response parts
			WithPluginInteractionContents(INTERACTION_PART_REQUEST, "application/protobuf", grpcInteraction)
		assert.NoError(t, err)

		// Start the gRPC mock server
		port, err := m.StartTransport("grpc", "127.0.0.1", 0, make(map[string][]interface{}))
		assert.NoError(t, err)
		defer m.CleanupMockServer(port)
		defer m.CleanupPlugins()

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
		assert.Nil(t, r)
		assert.ErrorContains(t, err, "not found")

		mismatches := m.MockServerMismatchedRequests(port)
		if len(mismatches) != 0 {
			assert.Len(t, mismatches, 0)
			t.Log(mismatches)
		}

		err = m.WritePactFileForServer(port, tmpPactFolder, true)
		assert.NoError(t, err)
	}
}
