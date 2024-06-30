//go:build plugin
// +build plugin

package v4

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsyncPluginTypeSystem(t *testing.T) {
	if runtime.GOOS != "windows" {

		// Sync - with plugin, but no transport
		// TODO: ExecuteTest has been disabled for now, because it's not very useful
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

		p, _ := NewAsynchronousPact(Config{
			Consumer: "asyncconsumer",
			Provider: "asyncprovider",
			PactDir:  "/tmp/",
		})

		// TODO: enable when there is a transport for async to test!
		err := p.AddAsynchronousMessage().
			Given("some state").
			ExpectsToReceive("some csv content").
			UsingPlugin(PluginConfig{
				Plugin:  "csv",
				Version: "0.0.6",
			}).
			WithContents(csvInteraction, "text/csv").
			// StartTransport("notarealtransport", "127.0.0.1", nil).
			ExecuteTest(t, func(m AsynchronousMessage) error {

				fmt.Println("Executing the CSV test", string(m.Contents))
				return nil
			})
		assert.NoError(t, err)
	}
}
