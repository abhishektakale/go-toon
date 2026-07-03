package toon

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ParseJSON reads JSON and keeps object key order. Numbers stay as json.Number.
func ParseJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return parseValue(dec)
}

// ParseJSONFast uses encoding/json directly. Key order is not preserved.
func ParseJSONFast(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	return normalizeDecodedJSON(raw), nil
}

// EncodeJSON is ParseJSON + Marshal.
func EncodeJSON(data []byte, opts ...EncodeOptions) ([]byte, error) {
	v, err := ParseJSON(data)
	if err != nil {
		return nil, err
	}
	return Marshal(v, opts...)
}

// EncodeJSONFast is the fast parse path when key order does not matter.
func EncodeJSONFast(data []byte, opts ...EncodeOptions) ([]byte, error) {
	v, err := ParseJSONFast(data)
	if err != nil {
		return nil, err
	}
	return Marshal(v, opts...)
}

func normalizeDecodedJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		obj := make(Object, 0, len(val))
		for key, item := range val {
			obj = append(obj, Field{Key: key, Value: normalizeDecodedJSON(item)})
		}
		return obj
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = normalizeDecodedJSON(item)
		}
		return out
	default:
		return val
	}
}

func parseValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := make(Object, 0, 4)
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("toon: expected object key, got %T", keyTok)
				}
				val, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				obj = append(obj, Field{Key: key, Value: val})
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return obj, nil
		case '[':
			arr := make([]any, 0, 4)
			for dec.More() {
				val, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return arr, nil
		default:
			return nil, fmt.Errorf("toon: unexpected delimiter %v", t)
		}
	case nil:
		return nil, nil
	case bool, string, json.Number:
		return t, nil
	default:
		return nil, fmt.Errorf("toon: unexpected token type %T", tok)
	}
}
