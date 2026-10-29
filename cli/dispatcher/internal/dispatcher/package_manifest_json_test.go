package dispatcher

import (
	"encoding/json"
	"testing"
)

func TestParseOrderedJSONObjectBytesRejectsIncompleteObjects(t *testing.T) {
	// Verifies empty, truncated, and value-less object input is rejected instead of yielding a partial object.
	for _, input := range []string{"", `{"a":1`, `{"a":}`} {
		object, err := parseOrderedJSONObjectBytes([]byte(input))
		if err == nil || len(object.keys) != 0 || len(object.values) != 0 {
			t.Fatalf("%q: expected an error and no partial object, got object=%+v err=%v", input, object, err)
		}
	}
}

func TestEmitOrderedJSONObjectRejectsInconsistentObjects(t *testing.T) {
	// Verifies emitting fails for a key without a value or with an unusable value instead of writing broken JSON.
	cases := []struct {
		name   string
		object orderedJSONObject
		want   string
	}{
		{name: "missing value", object: orderedJSONObject{keys: []string{"a"}, values: map[string]json.RawMessage{}}, want: `missing value for key "a"`},
		{name: "blank value", object: orderedJSONObject{keys: []string{"a"}, values: map[string]json.RawMessage{"a": json.RawMessage(" ")}}, want: "empty JSON value"},
		{name: "invalid array value", object: orderedJSONObject{keys: []string{"a"}, values: map[string]json.RawMessage{"a": json.RawMessage("[1,")}}, want: "unexpected end of JSON input"},
	}
	if _, err := emitOrderedJSONObject(orderedJSONObject{keys: []string{"a"}, values: map[string]json.RawMessage{"a": json.RawMessage(`{"b"`)}}, 0); err == nil {
		t.Fatal("a truncated nested object must not be emitted")
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := emitOrderedJSONObject(testCase.object, 0)
			if err == nil || err.Error() != testCase.want {
				t.Fatalf("expected %q, got %v", testCase.want, err)
			}
		})
	}
	if _, err := emitJSONRawArray([]json.RawMessage{json.RawMessage(" ")}, 0); err == nil || err.Error() != "empty JSON value" {
		t.Fatalf("expected the array element error, got %v", err)
	}
}
