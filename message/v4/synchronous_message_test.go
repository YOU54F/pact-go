package v4

import (
	"os"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v2/log"
)

func TestSyncTypeSystem(t *testing.T) {
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

	// Sync - no plugin
	_ = p.AddSynchronousMessage("some description").
		Given("some state").
		WithRequest(func(r *SynchronousMessageWithRequestBuilder) {
			r.WithJSONContent(map[string]string{"foo": "bar"})
			r.WithMetadata(map[string]string{})
		}).
		WithResponse(func(r *SynchronousMessageWithResponseBuilder) {
			r.WithJSONContent(map[string]string{"foo": "bar"})
			r.WithMetadata(map[string]string{})
		}).
		ExecuteTest(t, func(m SynchronousMessage) error {
			// In this scenario, we have no real transport, so we need to mock/handle both directions

			// e.g. MQ use case -> write to a queue, get a response from another queue

			// What is the user expected to do here?
			// m.Request. // inbound -> send to queue
			// Poll the queue
			// m.Response // response from queue

			// User consumes the request

			return nil
		})
}
