package toon_test

import (
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

// TestMarshalArrayOfMaps guards against a regression where map[string]any
// values (as produced by Decode) nested inside a slice were encoded as null
// instead of as objects.
func TestMarshalArrayOfMaps(t *testing.T) {
	v := []any{
		map[string]any{"id": 1.0, "name": "Alice"},
		map[string]any{"id": 2.0, "name": "Bob"},
	}
	out, err := toon.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := toon.Decode(out)
	if err != nil {
		t.Fatalf("decode %q: %v", out, err)
	}
	if !jsonEqual(decoded, v) {
		t.Fatalf("round trip mismatch: toon=%q", out)
	}
}

// TestMarshalMapFieldValue covers a map[string]any sitting as an object field
// value, which must encode as a nested object rather than null.
func TestMarshalMapFieldValue(t *testing.T) {
	v := toon.Object{{Key: "meta", Value: map[string]any{"a": 1.0}}}
	out, err := toon.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	want := "meta:\n  a: 1"
	if string(out) != want {
		t.Fatalf("got %q want %q", out, want)
	}
}
