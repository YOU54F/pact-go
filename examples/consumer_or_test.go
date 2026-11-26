//go:build consumer

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	AnyOf       = matchers.AnyOf
	LikeOrNull  = matchers.LikeOrNull
	RegexOrNull = matchers.RegexOrNull
)

// orRequest is what the consumer under test sends to the mock provider.
type orRequest struct {
	header string
	query  string
	body   string
}

// The example values: what the first alternative of each matcher describes.
var orExampleRequest = orRequest{
	header: "123",
	query:  "active",
	body:   `{"name": "Fred", "nickname": "Freddo", "address": {"street": "Main St", "number": 42}}`,
}

// TestConsumerV4AnyOf shows matching rules combined with OR: the mock server
// accepts a value that matches any one of the alternatives.
func TestConsumerV4AnyOf(t *testing.T) {
	require.NoError(t, log.SetLogLevel("INFO"))

	tests := map[string]struct {
		request orRequest
		wantErr bool
	}{
		"the example values match": {
			request: orExampleRequest,
		},
		"the other alternatives match": {
			request: orRequest{
				header: "abc",
				query:  "pending",
				body:   `{"name": null, "nickname": null, "address": null}`,
			},
		},
		"a string that is neither a string nor null fails": {
			request: orRequest{
				header: orExampleRequest.header,
				query:  orExampleRequest.query,
				body:   `{"name": 42, "nickname": "Freddo", "address": null}`,
			},
			wantErr: true,
		},
		"rules nested in the object example still apply": {
			request: orRequest{
				header: orExampleRequest.header,
				query:  orExampleRequest.query,
				body:   `{"name": "Fred", "nickname": "Freddo", "address": {"street": "Main St", "number": "42"}}`,
			},
			wantErr: true,
		},
		"a header matching no alternative fails": {
			request: orRequest{
				header: "ABC!",
				query:  orExampleRequest.query,
				body:   orExampleRequest.body,
			},
			wantErr: true,
		},
		"a query parameter matching no alternative fails": {
			request: orRequest{
				header: orExampleRequest.header,
				query:  "deleted",
				body:   orExampleRequest.body,
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pactDir := t.TempDir()
			err := anyOfInteraction(t, pactDir).ExecuteTest(t, orTest(tt.request))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assertPactCombinesWithOR(t, pactDir)
		})
	}
}

func anyOfInteraction(t *testing.T, pactDir string) *consumer.V4InteractionWithResponse {
	t.Helper()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV4AnyOfConsumer",
		Provider: "V4AnyOfProvider",
		Host:     "127.0.0.1",
		PactDir:  pactDir,
	})
	require.NoError(t, err)

	return mockProvider.
		AddInteraction().
		UponReceiving("A request using matchers combined with OR").
		WithRequest("POST", "/users", func(b *consumer.V4RequestBuilder) {
			b.
				Header("Content-Type", S("application/json")).
				Header("X-Id", AnyOf(Regex("123", `^\d+$`), Regex("abc", `^[a-z]+$`))).
				Query("status", AnyOf(Regex("active", "^active$"), Regex("pending", "^pending$"))).
				JSONBody(Map{
					"name":     LikeOrNull("Fred"),
					"nickname": RegexOrNull("Freddo", `^\w+$`),
					// Null or an object: the rules inside the example still
					// apply to its children when it is not null.
					"address": LikeOrNull(map[string]any{
						"street": Like("Main St"),
						"number": Integer(42),
					}),
				})
		}).
		WillRespondWith(200)
}

func orTest(r orRequest) func(config consumer.MockServerConfig) error {
	return func(config consumer.MockServerConfig) error {
		req := &http.Request{
			Method: http.MethodPost,
			URL: &url.URL{
				Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
				Scheme:   "http",
				Path:     "/users",
				RawQuery: url.Values{"status": {r.query}}.Encode(),
			},
			Body:   io.NopCloser(strings.NewReader(r.body)),
			Header: make(http.Header),
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Id", r.header)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("calling the mock provider: %w", err)
		}
		defer func() { _ = res.Body.Close() }()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status %d", res.StatusCode)
		}
		return nil
	}
}

// assertPactCombinesWithOR checks the pact file records the OR rule lists.
func assertPactCombinesWithOR(t *testing.T, pactDir string) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(pactDir, "PactGoV4AnyOfConsumer-V4AnyOfProvider.json")) //nolint:gosec // pactDir is the test's own temp dir
	require.NoError(t, err)

	var pact struct {
		Interactions []struct {
			Request struct {
				MatchingRules map[string]map[string]json.RawMessage `json:"matchingRules"`
			} `json:"request"`
		} `json:"interactions"`
	}
	require.NoError(t, json.Unmarshal(raw, &pact))
	require.Len(t, pact.Interactions, 1)
	rules := pact.Interactions[0].Request.MatchingRules

	assert.JSONEq(t, `{"combine": "OR", "matchers": [{"match": "type"}, {"match": "null"}]}`,
		string(rules["body"]["$.name"]))
	assert.JSONEq(t, `{"combine": "OR", "matchers": [{"match": "regex", "regex": "^\\w+$"}, {"match": "null"}]}`,
		string(rules["body"]["$.nickname"]))
	assert.JSONEq(t, `{"combine": "OR", "matchers": [{"match": "type"}, {"match": "null"}]}`,
		string(rules["body"]["$.address"]))
	assert.JSONEq(t, `{"combine": "AND", "matchers": [{"match": "integer"}]}`,
		string(rules["body"]["$.address.number"]))
	assert.JSONEq(t, `{"combine": "OR", "matchers": [{"match": "regex", "regex": "^\\d+$"}, {"match": "regex", "regex": "^[a-z]+$"}]}`,
		string(rules["header"]["X-Id"]))
	assert.JSONEq(t, `{"combine": "OR", "matchers": [{"match": "regex", "regex": "^active$"}, {"match": "regex", "regex": "^pending$"}]}`,
		string(rules["query"]["status"]))
}
