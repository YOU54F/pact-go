package native

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sendRawHeaderLines sends GET /headers to the mock server on port, writing
// each of lines as its own name: value field line. The request is written by
// hand because net/http merges repeated header lines into one before sending,
// and these tests depend on what reaches the mock server line by line.
func sendRawHeaderLines(t *testing.T, port int, name string, lines []string) int {
	t.Helper()

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	var dialer net.Dialer
	conn, err := dialer.DialContext(context.Background(), "tcp", address)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	var req strings.Builder
	fmt.Fprintf(&req, "GET /headers HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n", address)
	for _, line := range lines {
		fmt.Fprintf(&req, "%s: %s\r\n", name, line)
	}
	req.WriteString("\r\n")

	_, err = conn.Write([]byte(req.String()))
	require.NoError(t, err)

	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	return res.StatusCode
}

// regexHeader is a regex matching rule for a single header value, in the
// JSON form pactffi_with_header_v2 accepts.
func regexHeader(regex, example string) map[string]any {
	return map[string]any{
		"pact:matcher:type": "regex",
		"regex":             regex,
		"value":             example,
	}
}

// TestMockServer_RepeatedRequestHeaders checks how the mock server matches a
// header that is expected with several values (one value per index) against
// the field lines a client actually sends. RFC 9110 §5.3 treats repeated field
// lines as one line with the values joined by commas, so separate lines and a
// single comma-joined line should match the same expectation.
func TestMockServer_RepeatedRequestHeaders(t *testing.T) {
	linkA := `<https://a.example/x,y>; rel="a"`
	linkB := `<https://b.example/>; rel="b"`
	warns, noWarning := true, false

	tests := []struct {
		name     string
		header   string
		expected []any
		lines    []string
		match    bool
		// rfc9110Warning, when set, is whether the mock server should log the
		// pact-reference#584 deprecation warning for the request.
		rfc9110Warning *bool
	}{
		{
			name:     "H1 separate lines",
			header:   "X-Request",
			expected: []any{"a", "b"},
			lines:    []string{"a", "b"},
			match:    true,
		},
		{
			name:     "H2 one comma-joined line",
			header:   "X-Request",
			expected: []any{"a", "b"},
			lines:    []string{"a, b"},
			match:    true,
		},
		{
			name:     "H3 values containing commas on separate lines",
			header:   "X-Sunset",
			expected: []any{"Sat, 01 Jan 2028 00:00:00 GMT", "Sun, 02 Jan 2028 00:00:00 GMT"},
			lines:    []string{"Sat, 01 Jan 2028 00:00:00 GMT", "Sun, 02 Jan 2028 00:00:00 GMT"},
			match:    true,
		},
		{
			name:     "H4 singleton Content-Type sent as one comma-joined line",
			header:   "Content-Type",
			expected: []any{"text/plain", "text/html"},
			lines:    []string{"text/plain, text/html"},
			match:    false,
		},
		{
			name:     "H5 Cache-Control sent as one comma-joined line",
			header:   "Cache-Control",
			expected: []any{"no-cache", "no-store"},
			lines:    []string{"no-cache, no-store"},
			match:    true,
		},
		{
			name:   "H6a regex matcher per index",
			header: "X-Request",
			expected: []any{
				regexHeader(`^first-\d+$`, "first-1"),
				regexHeader(`^first-\d+$`, "first-2"),
			},
			lines:          []string{"first-42", "first-99"},
			match:          true,
			rfc9110Warning: &noWarning,
		},
		{
			// This only matches when the field lines are compared one at a time,
			// which pact-reference#584 deprecates with a warning.
			name:           "H6b quoted value only matching line by line",
			header:         "X-Request",
			expected:       []any{`"x , y"`, "z"},
			lines:          []string{`"x, y"`, "z"},
			match:          true,
			rfc9110Warning: &warns,
		},
		{
			name:     "H8a Link values with a comma in the URI on separate lines",
			header:   "Link",
			expected: []any{linkA, linkB},
			lines:    []string{linkA, linkB},
			match:    true,
		},
		{
			name:     "H8b Link values with a comma in the URI on one comma-joined line",
			header:   "Link",
			expected: []any{linkA, linkB},
			lines:    []string{linkA + ", " + linkB},
			match:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHTTPPact("consumer", "provider")
			m.NewInteraction(tt.name).
				UponReceiving(tt.name).
				WithRequest("GET", "/headers").
				WithRequestHeaders(map[string][]any{tt.header: tt.expected}).
				WithStatus(200)

			port, err := m.Start("127.0.0.1:0", false)
			require.NoError(t, err)
			defer m.CleanupMockServer(port)

			status := sendRawHeaderLines(t, port, tt.header, tt.lines)
			mismatches := m.MockServerMismatchedRequests(port)

			if tt.match {
				assert.Equal(t, http.StatusOK, status)
				assert.Empty(t, mismatches)
			} else {
				assert.NotEqual(t, http.StatusOK, status)
				assert.NotEmpty(t, mismatches)
			}

			if tt.rfc9110Warning != nil {
				logs := m.MockServerLogs(port)
				if *tt.rfc9110Warning {
					assert.Contains(t, logs, "RFC 9110 §5.3")
				} else {
					assert.NotContains(t, logs, "RFC 9110 §5.3")
				}
			}
		})
	}
}
