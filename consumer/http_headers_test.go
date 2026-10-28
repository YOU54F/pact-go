package consumer

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repeatedHeaderCase is a request header expected with one value per index,
// and the field lines the client sends for it.
type repeatedHeaderCase struct {
	name     string
	header   string
	expected []matchers.Matcher
	lines    []string
	match    bool
	// rfc9110Warning, when set, is whether the mock server should log the
	// pact-reference#584 deprecation warning for the request.
	rfc9110Warning *bool
}

// rfc9110WarningText is part of the deprecation warning pact_ffi logs when
// repeated header values only match when their field lines are compared one
// at a time (pact-reference#584).
const rfc9110WarningText = "RFC 9110 §5.3"

// repeatedHeaderCases covers how the mock server matches repeated request
// headers. RFC 9110 §5.3 treats repeated field lines as one line with the
// values joined by commas, so separate lines and a single comma-joined line
// should match the same expectation.
func repeatedHeaderCases() []repeatedHeaderCase {
	linkA := `<https://a.example/x,y>; rel="a"`
	linkB := `<https://b.example/>; rel="b"`
	warns, noWarning := true, false

	return []repeatedHeaderCase{
		{
			name:     "H1 separate lines",
			header:   "X-Request",
			expected: []matchers.Matcher{S("a"), S("b")},
			lines:    []string{"a", "b"},
			match:    true,
		},
		{
			name:     "H2 one comma-joined line",
			header:   "X-Request",
			expected: []matchers.Matcher{S("a"), S("b")},
			lines:    []string{"a, b"},
			match:    true,
		},
		{
			name:     "H3 values containing commas on separate lines",
			header:   "X-Sunset",
			expected: []matchers.Matcher{S("Sat, 01 Jan 2028 00:00:00 GMT"), S("Sun, 02 Jan 2028 00:00:00 GMT")},
			lines:    []string{"Sat, 01 Jan 2028 00:00:00 GMT", "Sun, 02 Jan 2028 00:00:00 GMT"},
			match:    true,
		},
		{
			name:     "H4 singleton Content-Type sent as one comma-joined line",
			header:   "Content-Type",
			expected: []matchers.Matcher{S("text/plain"), S("text/html")},
			lines:    []string{"text/plain, text/html"},
			match:    false,
		},
		{
			name:     "H5 Cache-Control sent as one comma-joined line",
			header:   "Cache-Control",
			expected: []matchers.Matcher{S("no-cache"), S("no-store")},
			lines:    []string{"no-cache, no-store"},
			match:    true,
		},
		{
			name:           "H6a regex matcher per index",
			header:         "X-Request",
			expected:       []matchers.Matcher{Regex("first-1", `^first-\d+$`), Regex("first-2", `^first-\d+$`)},
			lines:          []string{"first-42", "first-99"},
			match:          true,
			rfc9110Warning: &noWarning,
		},
		{
			// This only matches when the field lines are compared one at a time,
			// which pact-reference#584 deprecates with a warning.
			name:           "H6b quoted value only matching line by line",
			header:         "X-Request",
			expected:       []matchers.Matcher{S(`"x , y"`), S("z")},
			lines:          []string{`"x, y"`, "z"},
			match:          true,
			rfc9110Warning: &warns,
		},
		{
			name:     "H8a Link values with a comma in the URI on separate lines",
			header:   "Link",
			expected: []matchers.Matcher{S(linkA), S(linkB)},
			lines:    []string{linkA, linkB},
			match:    true,
		},
		{
			name:     "H8b Link values with a comma in the URI on one comma-joined line",
			header:   "Link",
			expected: []matchers.Matcher{S(linkA), S(linkB)},
			lines:    []string{linkA + ", " + linkB},
			match:    true,
		},
	}
}

// assertRFC9110Warning checks whether logs hold the pact-reference#584
// deprecation warning, when tt expects one way or the other.
func assertRFC9110Warning(t *testing.T, tt repeatedHeaderCase, logs string) {
	t.Helper()
	if tt.rfc9110Warning == nil {
		return
	}
	if *tt.rfc9110Warning {
		assert.Contains(t, logs, rfc9110WarningText)
	} else {
		assert.NotContains(t, logs, rfc9110WarningText)
	}
}

// sendRawHeaderLines sends GET /headers to the mock server, writing each of
// lines as its own name: value field line. The request is written by hand
// because net/http merges repeated header lines into one before sending.
func sendRawHeaderLines(msc MockServerConfig, name string, lines []string) (int, error) {
	address := net.JoinHostPort(msc.Host, strconv.Itoa(msc.Port))
	var dialer net.Dialer
	conn, err := dialer.DialContext(context.Background(), "tcp", address)
	if err != nil {
		return 0, fmt.Errorf("connecting to the mock server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	var req strings.Builder
	fmt.Fprintf(&req, "GET /headers HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n", address)
	for _, line := range lines {
		fmt.Fprintf(&req, "%s: %s\r\n", name, line)
	}
	req.WriteString("\r\n")

	_, err = conn.Write([]byte(req.String()))
	if err != nil {
		return 0, fmt.Errorf("writing the request: %w", err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return 0, fmt.Errorf("reading the response: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	return res.StatusCode, nil
}

func TestV3HTTPRepeatedRequestHeaders(t *testing.T) {
	for _, tt := range repeatedHeaderCases() {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewV3Pact(MockHTTPProviderConfig{
				Consumer: "consumer",
				Provider: "provider",
				PactDir:  t.TempDir(),
			})
			require.NoError(t, err)

			var status int
			err = p.AddInteraction().
				UponReceiving(tt.name).
				WithRequest("GET", "/headers", func(b *V3RequestBuilder) {
					b.Header(tt.header, tt.expected...)
				}).
				WillRespondWith(200).
				ExecuteTest(t, func(msc MockServerConfig) error {
					var sendErr error
					status, sendErr = sendRawHeaderLines(msc, tt.header, tt.lines)
					return sendErr
				})

			if tt.match {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, status)
			} else {
				require.Error(t, err)
				assert.NotEqual(t, http.StatusOK, status)
			}
			assertRFC9110Warning(t, tt, p.MockServerLogs())
		})
	}
}

func TestV4HTTPRepeatedRequestHeaders(t *testing.T) {
	for _, tt := range repeatedHeaderCases() {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewV4Pact(MockHTTPProviderConfig{
				Consumer: "consumer",
				Provider: "provider",
				PactDir:  t.TempDir(),
			})
			require.NoError(t, err)

			var status int
			err = p.AddInteraction().
				UponReceiving(tt.name).
				WithRequest("GET", "/headers", func(b *V4RequestBuilder) {
					b.Header(tt.header, tt.expected...)
				}).
				WillRespondWith(200).
				ExecuteTest(t, func(msc MockServerConfig) error {
					var sendErr error
					status, sendErr = sendRawHeaderLines(msc, tt.header, tt.lines)
					return sendErr
				})

			if tt.match {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, status)
			} else {
				require.Error(t, err)
				assert.NotEqual(t, http.StatusOK, status)
			}
			assertRFC9110Warning(t, tt, p.MockServerLogs())
		})
	}
}
