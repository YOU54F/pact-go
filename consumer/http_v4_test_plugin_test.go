//go:build plugin
// +build plugin

package consumer

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestHttpV4PluginTypeSystem(t *testing.T) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "TRACE"
	}
	_ = log.SetLogLevel(logutils.LogLevel(logLevel))

	p, err := NewV4Pact(MockHTTPProviderConfig{
		Consumer: "consumer",
		Provider: "provider",
	})
	assert.NoError(t, err)

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

	err = p.AddInteraction().
		Given("some state").
		UponReceiving("some scenario").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithRequest("GET", "/").
		// WithRequest("GET", "/", func(b *V4InteractionWithPluginRequestBuilder) {
		// 	b.PluginContents("application/protobufs", "")
		// 	// TODO:
		// }).
		WillRespondWith(200, func(b *V4InteractionWithPluginResponseBuilder) {
			b.
				Header("Content-Type", S("application/protobufs")).
				PluginContents("application/protobufs", `
					{
						"pact:proto": "`+path+`",
						"pact:message-type": "InitPluginRequest"
					}
				`)
		}).
		ExecuteTest(t, func(msc MockServerConfig) error {
			// <- normally run the actually test here.

			return nil
		})
	assert.Error(t, err)

}
