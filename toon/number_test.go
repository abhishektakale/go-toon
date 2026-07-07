package toon_test

import (
	"math"
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

// TestEncodeJSONNumberFormatting locks the integer/float formatting boundaries
// that formatJSONNumber and formatFloat64 produce.
func TestEncodeJSONNumberFormatting(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"n":0}`, "n: 0"},
		{`{"n":-0}`, "n: 0"},
		{`{"n":42}`, "n: 42"},
		{`{"n":-7}`, "n: -7"},
		{`{"n":9223372036854775807}`, "n: 9223372036854775807"}, // max int64
		{`{"n":1.5}`, "n: 1.5"},
		{`{"n":1.0}`, "n: 1"},
		{`{"n":1e2}`, "n: 100"},
		{`{"n":1.25e2}`, "n: 125"},
		{`{"n":12.50}`, "n: 12.5"},
		{`{"n":1e21}`, "n: 1e+21"},
		{`{"n":0.000001}`, "n: 0.000001"},
		{`{"n":1e-7}`, "n: 1e-07"},
	}
	for _, tc := range cases {
		out, err := toon.EncodeJSON([]byte(tc.in))
		if err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		if string(out) != tc.want {
			t.Errorf("EncodeJSON(%s) = %q, want %q", tc.in, out, tc.want)
		}
	}
}

// TestDecodeNegativeZero verifies that "-0" decodes to positive zero so it
// re-serializes as 0 rather than -0.
func TestDecodeNegativeZero(t *testing.T) {
	v, err := toon.Decode([]byte("n: -0"))
	if err != nil {
		t.Fatal(err)
	}
	m := v.(map[string]any)
	f, ok := m["n"].(float64)
	if !ok || f != 0 || math.Signbit(f) {
		t.Fatalf("n = %#v, want +0", m["n"])
	}

	out, err := toon.DecodeToJSON([]byte("n: -0"))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"n":0}` {
		t.Fatalf("DecodeToJSON = %s, want {\"n\":0}", out)
	}
}
