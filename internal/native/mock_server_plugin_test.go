//go:build plugin
// +build plugin

package native

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPluginInteraction(t *testing.T) {

	tmpPactFolder, err := os.MkdirTemp("", "pact-go")
	assert.NoError(t, err)
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "TRACE"
	}
	_ = log.SetLogLevel(logutils.LogLevel(logLevel))
	m := NewHTTPPact("test-plugin-consumer", "test-plugin-provider")

	// Protobuf plugin test
	_ = m.UsingPlugin("protobuf", "0.3.15")
	m.WithSpecificationVersion(SPECIFICATION_VERSION_V4)

	i := m.NewInteraction("some plugin interaction")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

	protobufInteraction := `{
			"pact:proto": "` + path + `",
			"pact:message-type": "InitPluginRequest",
			"pact:content-type": "application/protobuf",
			"implementation": "notEmpty('pact-go-driver')",
			"version": "matching(semver, '0.0.0')"
		}`

	err = i.UponReceiving("some interaction").
		Given("plugin state").
		WithRequest("GET", "/protobuf").
		WithStatus(200).
		WithPluginInteractionContents(INTERACTION_PART_RESPONSE, "application/protobuf", protobufInteraction)
	assert.NoError(t, err)

	port, err := m.Start("0.0.0.0:0", false)
	assert.NoError(t, err)
	defer m.CleanupMockServer(port)
	defer m.CleanupPlugins()

	res, err := http.Get(fmt.Sprintf("http://0.0.0.0:%d/protobuf", port))
	assert.NoError(t, err)

	bytes, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	initPluginRequest := &InitPluginRequest{}
	err = proto.Unmarshal(bytes, initPluginRequest)
	assert.NoError(t, err)

	assert.Equal(t, "pact-go-driver", initPluginRequest.Implementation)
	assert.Equal(t, "0.0.0", initPluginRequest.Version)

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 0 {
		assert.Len(t, mismatches, 0)
		t.Log(mismatches)
	}

	err = m.WritePactFile(port, tmpPactFolder)
	assert.NoError(t, err)

}
