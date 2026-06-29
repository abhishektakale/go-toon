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

// EncodeJSON is ParseJSON + Marshal.
func EncodeJSON(data []byte, opts ...EncodeOptions) ([]byte, error) {
	v, err := ParseJSON(data)
	if err != nil {
		return nil, err
	}
	return Marshal(v, opts...)
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
