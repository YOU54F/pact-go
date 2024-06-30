//go:build plugin
// +build plugin

package v4

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestSyncPluginTypeSystem(t *testing.T) {

	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "TRACE"
	}
	_ = log.SetLogLevel(logutils.LogLevel(logLevel))

	// Sync - with plugin, but no transport
	csvInteraction := `{
		"request.path": "/reports/report002.csv",
		"response.status": "200",
		"response.contents": {
			"pact:content-type": "text/csv",
			"csvHeaders": true,
			"column:Name": "matching(type,'Name')",
			"column:Number": "matching(number,100)",
			"column:Date": "matching(datetime, 'yyyy-MM-dd','2000-01-01')"
		}
	}`

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	_ = p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "csv",
			Version: "0.0.6",
		}).
		WithContents(csvInteraction, "text/csv").
		ExecuteTest(t, func(m SynchronousMessage) error {
			fmt.Println("Executing the CSV test")
			return nil
		})

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})

	// Sync - with plugin + transport (pass)
	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/../../internal/native/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

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

	err := p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return nil
		})

	assert.Error(t, err)

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})

	// Sync - with plugin + transport (fail)
	err = p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return errors.New("bad thing")
		})

	assert.Error(t, err)

}
