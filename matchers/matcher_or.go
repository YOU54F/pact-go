package matchers

import (
	"encoding/json"
	"errors"
	"fmt"
)

// combineOR is the pact:matcher:combine value that makes at least one of the
// rules in a pact:matcher:type array apply (rather than all of them, AND).
const combineOR = "OR"

// matcherTypeKey is the integration JSON key naming a matcher's kind.
const matcherTypeKey = "pact:matcher:type"

// errNoMatchers is returned when AnyOf is marshalled without any alternatives.
var errNoMatchers = errors.New("AnyOf requires at least one matcher")

// errNotARule is returned when an AnyOf alternative does not encode a
// matching rule (for example a plain String or a StructMatcher).
var errNotARule = errors.New("AnyOf alternative is not a matching rule")

// anyOf combines several matchers for the same path with OR logic.
type anyOf struct {
	Matchers []Matcher
}

// GetValue returns the example of the first alternative.
func (a anyOf) GetValue() any {
	if len(a.Matchers) == 0 {
		return nil
	}
	return a.Matchers[0].GetValue()
}

// MarshalJSON encodes a as a single integration JSON matcher whose
// pact:matcher:type is the flattened list of alternatives, combined with OR,
// and whose value is the example of the first alternative.
//
// Fields here are emitted in a fixed order, keeping pact:matcher:type first:
// the key order is contract-visible in pact files another party consumes.
func (a anyOf) MarshalJSON() ([]byte, error) {
	if len(a.Matchers) == 0 {
		return nil, errNoMatchers
	}

	rules, err := a.rules()
	if err != nil {
		return nil, err
	}

	value, err := exampleOf(a.Matchers[0])
	if err != nil {
		return nil, err
	}

	//nolint:wrapcheck // encoding/json wraps a MarshalJSON failure in a
	// *json.MarshalerError recording the type and the original error, so an extra
	// prefix here would be doubled up in the message the caller finally sees.
	return json.Marshal(struct {
		Type    []map[string]json.RawMessage `json:"pact:matcher:type"`
		Combine string                       `json:"pact:matcher:combine"`
		Value   json.RawMessage              `json:"value"`
	}{rules, combineOR, value})
}

func (a anyOf) isMatcher() {}

// rules returns the matching rule of each alternative, without its example,
// flattening any nested AnyOf so that all the alternatives end up in the one
// list (the FFI replaces, rather than merges, an OR group nested in the value).
func (a anyOf) rules() ([]map[string]json.RawMessage, error) {
	rules := make([]map[string]json.RawMessage, 0, len(a.Matchers))
	for idx, m := range a.Matchers {
		if nested, ok := m.(anyOf); ok {
			nestedRules, err := nested.rules()
			if err != nil {
				return nil, err
			}
			rules = append(rules, nestedRules...)
			continue
		}

		fields, err := marshalToFields(m)
		if err != nil {
			return nil, err
		}
		if _, ok := fields[matcherTypeKey]; !ok {
			return nil, fmt.Errorf("%w: alternative %d (%T)", errNotARule, idx, m)
		}

		// The example and specification belong to the combined matcher, not to
		// each rule in the list.
		delete(fields, "value")
		delete(fields, "pact:specification")
		delete(fields, "specification")
		rules = append(rules, fields)
	}
	return rules, nil
}

// exampleOf returns the JSON example of m, keeping any matchers nested inside
// it so that rules for child paths still apply.
func exampleOf(m Matcher) (json.RawMessage, error) {
	if nested, ok := m.(anyOf); ok {
		if len(nested.Matchers) == 0 {
			return nil, errNoMatchers
		}
		return exampleOf(nested.Matchers[0])
	}

	fields, err := marshalToFields(m)
	if err != nil {
		return nil, err
	}
	if value, ok := fields["value"]; ok {
		return value, nil
	}

	//nolint:wrapcheck // see MarshalJSON above.
	return json.Marshal(m.GetValue())
}

// marshalToFields encodes m and splits the resulting JSON object into its
// top level fields. Matchers that do not encode to an object yield no fields.
func marshalToFields(m Matcher) (map[string]json.RawMessage, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		//nolint:wrapcheck // see MarshalJSON above.
		return nil, err
	}

	fields := map[string]json.RawMessage{}
	if len(raw) == 0 || raw[0] != '{' {
		return fields, nil
	}
	err = json.Unmarshal(raw, &fields)
	if err != nil {
		//nolint:wrapcheck // see MarshalJSON above.
		return nil, err
	}
	return fields, nil
}

// AnyOf matches a value if at least one of the given matchers matches it
// (the matching rules are combined with OR).
//
// The first matcher provides the example used in the consumer test, including
// any matchers nested inside an object example. The other alternatives only
// contribute their matching rule: they are recorded in the pact file and
// applied during provider verification, but are never exercised against your
// consumer code.
//
// Each alternative must be a matcher that encodes a matching rule (Like,
// Regex, Integer, Null, ...), not a plain String or StructMatcher. Nested
// AnyOf matchers are flattened into the one list of alternatives.
//
// Requires a pact_ffi version that supports pact:matcher:combine
// (pact-foundation/pact-reference#601).
//
// TODO: replace with the released pact_ffi version once available.
func AnyOf(matchers ...Matcher) Matcher {
	return anyOf{Matchers: matchers}
}

// OrNull matches a value that matches m, or is null. The example is m's
// example.
func OrNull(m Matcher) Matcher {
	return AnyOf(m, Null{})
}

// LikeOrNull matches a value of the same type as content, or null.
func LikeOrNull(content any) Matcher {
	return OrNull(Like(content))
}

// IntegerOrNull matches an integer, or null.
func IntegerOrNull(example int) Matcher {
	return OrNull(Integer(example))
}

// DecimalOrNull matches a decimal, or null.
func DecimalOrNull(example float64) Matcher {
	return OrNull(Decimal(example))
}

// RegexOrNull matches a string matching regex, or null.
func RegexOrNull(example string, regex string) Matcher {
	return OrNull(Regex(example, regex))
}
