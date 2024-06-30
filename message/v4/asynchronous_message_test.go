package v4

import (
	"os"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestAsyncTypeSystem(t *testing.T) {
	p, _ := NewAsynchronousPact(Config{
		Consumer: "asyncconsumer",
		Provider: "asyncprovider",
		PactDir:  "/tmp/",
	})
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "TRACE"
	}
	_ = log.SetLogLevel(logutils.LogLevel(logLevel))

	type foo struct {
		Foo string `json:"foo"`
	}

	// Sync - no plugin
	err := p.AddAsynchronousMessage().
		Given("some state").
		Given("another state").
		ExpectsToReceive("an important json message").
		WithJSONContent(map[string]string{
			"foo": "bar",
		}).
		AsType(&foo{}).
		ConsumedBy(func(mc AsynchronousMessage) error {
			fooMessage := mc.Body.(*foo)
			assert.Equal(t, "bar", fooMessage.Foo)
			return nil
		}).
		Verify(t)

	assert.NoError(t, err)

}
