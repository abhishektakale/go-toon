package toon_test

import (
	"encoding/json"
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

// TestRoundTripShapes checks JSON -> TOON -> JSON across the shapes the single
// tabular round-trip test does not cover: nested arrays, key folding paired
// with path expansion, and non-comma delimiters.
func TestRoundTripShapes(t *testing.T) {
	cases := []struct {
		name   string
		json   string
		encode toon.EncodeOptions
		decode toon.DecodeOptions
	}{
		{
			name: "nested primitive arrays",
			json: `{"matrix":[[1,2,3],[4,5,6]],"tags":["a","b"]}`,
		},
		{
			name: "nested objects and tabular arrays",
			json: `{"user":{"id":1,"roles":["admin","ops"]},"items":[{"sku":"A1","qty":2},{"sku":"B2","qty":3}]}`,
		},
		{
			name:   "key folding paired with path expansion",
			json:   `{"a":{"b":{"c":1}},"x":2}`,
			encode: toon.EncodeOptions{KeyFolding: toon.KeyFoldingSafe},
			decode: toon.DecodeOptions{ExpandPaths: "safe"},
		},
		{
			name:   "tab delimiter",
			json:   `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`,
			encode: toon.EncodeOptions{Delimiter: toon.DelimiterTab},
			decode: toon.DecodeOptions{Delimiter: toon.DelimiterTab},
		},
		{
			name:   "pipe delimiter",
			json:   `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`,
			encode: toon.EncodeOptions{Delimiter: toon.DelimiterPipe},
			decode: toon.DecodeOptions{Delimiter: toon.DelimiterPipe},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.decode.Strict = true

			toonBytes, err := toon.EncodeJSON([]byte(tc.json), tc.encode)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			decoded, err := toon.Decode(toonBytes, tc.decode)
			if err != nil {
				t.Fatalf("decode %q: %v", toonBytes, err)
			}

			var want any
			if err := json.Unmarshal([]byte(tc.json), &want); err != nil {
				t.Fatalf("parse expected json: %v", err)
			}
			if !jsonEqual(decoded, want) {
				got, _ := json.Marshal(decoded)
				t.Fatalf("round trip mismatch\ntoon:\n%s\ngot:  %s\nwant: %s", toonBytes, got, tc.json)
			}
		})
	}
}
