package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// orderedJSONObject preserves the original key order of a JSON object so a
// rewrite of Packages/manifest.json produces a minimal diff.
type orderedJSONObject struct {
	keys   []string
	values map[string]json.RawMessage
}

func parseOrderedJSONObjectBytes(data []byte) (orderedJSONObject, error) {
	trimmed := bytes.TrimSpace(data)
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	token, err := dec.Token()
	if err != nil {
		return orderedJSONObject{}, err
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '{' {
		return orderedJSONObject{}, fmt.Errorf("expected JSON object")
	}

	object := orderedJSONObject{
		keys:   []string{},
		values: map[string]json.RawMessage{},
	}
	for dec.More() {
		keyToken, keyErr := dec.Token()
		if keyErr != nil {
			return orderedJSONObject{}, keyErr
		}
		key, keyOK := keyToken.(string)
		if !keyOK {
			return orderedJSONObject{}, fmt.Errorf("expected object key")
		}
		var raw json.RawMessage
		if decodeErr := dec.Decode(&raw); decodeErr != nil {
			return orderedJSONObject{}, decodeErr
		}
		object.keys = append(object.keys, key)
		object.values[key] = raw
	}
	closing, err := dec.Token()
	if err != nil {
		return orderedJSONObject{}, err
	}
	if closingDelim, ok := closing.(json.Delim); !ok || closingDelim != '}' {
		return orderedJSONObject{}, fmt.Errorf("expected end of JSON object")
	}
	if err := ensureJSONDecoderFullyConsumed(dec); err != nil {
		return orderedJSONObject{}, err
	}
	return object, nil
}

// ensureJSONDecoderFullyConsumed rejects trailing tokens after a complete JSON
// value so malformed manifests with garbage after the closing delimiter fail
// instead of being silently rewritten without that garbage.
func ensureJSONDecoderFullyConsumed(dec *json.Decoder) error {
	token, err := dec.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("unexpected trailing JSON token: %v", token)
}

func emitOrderedJSONObject(object orderedJSONObject, depth int) (json.RawMessage, error) {
	indent := strings.Repeat("  ", depth)
	innerIndent := strings.Repeat("  ", depth+1)
	var builder strings.Builder
	builder.WriteString("{\n")
	for index, key := range object.keys {
		raw, ok := object.values[key]
		if !ok {
			return nil, fmt.Errorf("missing value for key %q", key)
		}
		encodedKey, _ := json.Marshal(key)
		formattedValue, err := formatJSONRawForEmit(raw, depth+1)
		if err != nil {
			return nil, err
		}
		builder.WriteString(innerIndent)
		builder.Write(encodedKey)
		builder.WriteString(": ")
		builder.WriteString(formattedValue)
		if index < len(object.keys)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	builder.WriteString(indent)
	builder.WriteString("}")
	return json.RawMessage(builder.String()), nil
}

func emitJSONRawArray(elements []json.RawMessage, depth int) (json.RawMessage, error) {
	indent := strings.Repeat("  ", depth)
	innerIndent := strings.Repeat("  ", depth+1)
	var builder strings.Builder
	builder.WriteString("[\n")
	for index, element := range elements {
		formatted, err := formatJSONRawForEmit(element, depth+1)
		if err != nil {
			return nil, err
		}
		builder.WriteString(innerIndent)
		builder.WriteString(formatted)
		if index < len(elements)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	builder.WriteString(indent)
	builder.WriteString("]")
	return json.RawMessage(builder.String()), nil
}

func formatJSONRawForEmit(raw json.RawMessage, depth int) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "", fmt.Errorf("empty JSON value")
	}
	switch trimmed[0] {
	case '{':
		object, err := parseOrderedJSONObjectBytes(trimmed)
		if err != nil {
			return "", err
		}
		emitted, err := emitOrderedJSONObject(object, depth)
		if err != nil {
			return "", err
		}
		return string(emitted), nil
	case '[':
		// Pretty-print arrays (including scopes) with stable formatting.
		var values []json.RawMessage
		if err := json.Unmarshal(trimmed, &values); err != nil {
			return "", err
		}
		// Distinguish object-array (scopedRegistries elements) from string-array (scopes).
		if len(values) > 0 {
			first := bytes.TrimSpace(values[0])
			if len(first) > 0 && first[0] == '{' {
				emitted, err := emitJSONRawArray(values, depth)
				if err != nil {
					return "", err
				}
				return string(emitted), nil
			}
		}
		indent := strings.Repeat("  ", depth)
		innerIndent := strings.Repeat("  ", depth+1)
		var builder strings.Builder
		builder.WriteString("[\n")
		for index, value := range values {
			builder.WriteString(innerIndent)
			builder.Write(bytes.TrimSpace(value))
			if index < len(values)-1 {
				builder.WriteString(",")
			}
			builder.WriteString("\n")
		}
		builder.WriteString(indent)
		builder.WriteString("]")
		return builder.String(), nil
	default:
		return string(trimmed), nil
	}
}
