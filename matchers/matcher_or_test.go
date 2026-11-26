package matchers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func marshalString(t *testing.T, m Matcher) string {
	t.Helper()
	b, err := json.Marshal(m)
	require.NoError(t, err)
	return string(b)
}

func TestMatcher_AnyOfTypeOrNull(t *testing.T) {
	assert.JSONEq(t, `{
		"pact:matcher:type": [
			{"pact:matcher:type": "type"},
			{"pact:matcher:type": "null"}
		],
		"pact:matcher:combine": "OR",
		"value": "Fred"
	}`, marshalString(t, AnyOf(Like("Fred"), Null{})))
}

func TestMatcher_AnyOfKeepsRuleAttributes(t *testing.T) {
	assert.JSONEq(t, `{
		"pact:matcher:type": [
			{"pact:matcher:type": "regex", "regex": "^\\d+$"},
			{"pact:matcher:type": "integer"},
			{"pact:matcher:type": "type", "min": 1}
		],
		"pact:matcher:combine": "OR",
		"value": "100"
	}`, marshalString(t, AnyOf(Regex("100", `^\d+$`), Integer(5), EachLike("x", 1))))
}

func TestMatcher_AnyOfKeyOrder(t *testing.T) {
	// pact:matcher:type first, as with the other matchers.
	assert.Regexp(t, `^\{"pact:matcher:type":\[`, marshalString(t, AnyOf(Like("Fred"), Null{})))
}

func TestMatcher_AnyOfFlattensNested(t *testing.T) {
	assert.JSONEq(t, `{
		"pact:matcher:type": [
			{"pact:matcher:type": "integer"},
			{"pact:matcher:type": "null"},
			{"pact:matcher:type": "regex", "regex": "^\\d+$"}
		],
		"pact:matcher:combine": "OR",
		"value": 42
	}`, marshalString(t, AnyOf(IntegerOrNull(42), Regex("1", `^\d+$`))))
}

func TestMatcher_AnyOfNullOrObjectKeepsNestedMatchers(t *testing.T) {
	m := LikeOrNull(map[string]any{
		"street": "Main St",
		"number": Integer(42),
	})
	assert.JSONEq(t, `{
		"pact:matcher:type": [
			{"pact:matcher:type": "type"},
			{"pact:matcher:type": "null"}
		],
		"pact:matcher:combine": "OR",
		"value": {
			"street": "Main St",
			"number": {"specification": "3.0.0", "pact:matcher:type": "integer", "value": 42}
		}
	}`, marshalString(t, m))
}

func TestMatcher_AnyOfNullFirstHasNullExample(t *testing.T) {
	assert.JSONEq(t, `{
		"pact:matcher:type": [
			{"pact:matcher:type": "null"},
			{"pact:matcher:type": "type"}
		],
		"pact:matcher:combine": "OR",
		"value": null
	}`, marshalString(t, AnyOf(Null{}, Like("Fred"))))
	assert.Nil(t, AnyOf(Null{}, Like("Fred")).GetValue())
}

func TestMatcher_AnyOfInsideBody(t *testing.T) {
	body := map[string]any{"name": LikeOrNull("Fred")}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name": {
		"pact:matcher:type": [{"pact:matcher:type": "type"}, {"pact:matcher:type": "null"}],
		"pact:matcher:combine": "OR",
		"value": "Fred"
	}}`, string(b))
}

func TestMatcher_OrNullHelpers(t *testing.T) {
	tests := map[string]struct {
		matcher Matcher
		rule    string
		value   string
	}{
		"LikeOrNull":    {LikeOrNull("Fred"), `{"pact:matcher:type":"type"}`, `"Fred"`},
		"IntegerOrNull": {IntegerOrNull(7), `{"pact:matcher:type":"integer"}`, `7`},
		"DecimalOrNull": {DecimalOrNull(1.5), `{"pact:matcher:type":"decimal"}`, `1.5`},
		"RegexOrNull":   {RegexOrNull("abc", `\w+`), `{"pact:matcher:type":"regex","regex":"\\w+"}`, `"abc"`},
		"OrNull":        {OrNull(Includes("ell")), `{"pact:matcher:type":"include"}`, `"ell"`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.JSONEq(t, `{
				"pact:matcher:type": [`+tt.rule+`, {"pact:matcher:type": "null"}],
				"pact:matcher:combine": "OR",
				"value": `+tt.value+`
			}`, marshalString(t, tt.matcher))
		})
	}
}

func TestMatcher_AnyOfGetValue(t *testing.T) {
	assert.Equal(t, "Fred", AnyOf(Like("Fred"), Null{}).GetValue())
	assert.Equal(t, 42, IntegerOrNull(42).GetValue())
	assert.Nil(t, AnyOf().GetValue())
}

func TestMatcher_AnyOfErrors(t *testing.T) {
	_, err := json.Marshal(AnyOf())
	require.ErrorIs(t, err, errNoMatchers)

	_, err = json.Marshal(AnyOf(Like("x"), String("y")))
	require.ErrorIs(t, err, errNotARule)

	_, err = json.Marshal(AnyOf(Like("x"), StructMatcher{"a": Like(1)}))
	require.ErrorIs(t, err, errNotARule)
}
