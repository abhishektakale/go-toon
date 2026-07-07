package toon_test

import (
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

// FuzzDecodeNoPanic asserts the decoder returns an error rather than panicking
// on arbitrary input. Run with: go test ./toon -run=^$ -fuzz=FuzzDecodeNoPanic
func FuzzDecodeNoPanic(f *testing.F) {
	seeds := []string{
		"",
		"[]",
		"a: 1",
		"x: true",
		"users[2]{id,name}:\n  1,Alice\n  2,Bob",
		"matrix[2]:\n  - [3]: 1,2,3\n  - [3]: 4,5,6",
		"nums[3]: 1,2,3",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = toon.Decode([]byte(s))
	})
}

// FuzzRoundTrip checks that decoding is idempotent through a Marshal: for any
// input the decoder accepts, re-encoding and decoding again yields an equal
// value. Run with: go test ./toon -run=^$ -fuzz=FuzzRoundTrip
func FuzzRoundTrip(f *testing.F) {
	seeds := []string{
		"a: 1",
		"x: true\ny: false\nz: null",
		"users[2]{id,name}:\n  1,Alice\n  2,Bob",
		"nums[3]: 1,2,3",
		"matrix[2]:\n  - [3]: 1,2,3\n  - [3]: 4,5,6",
		"user:\n  id: 1\n  name: Ada",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := toon.Decode([]byte(s))
		if err != nil {
			return // only exercise inputs the decoder accepts
		}
		out, err := toon.Marshal(v)
		if err != nil {
			t.Fatalf("marshal decoded value: %v", err)
		}
		v2, err := toon.Decode(out)
		if err != nil {
			t.Fatalf("re-decode %q: %v", out, err)
		}
		if !jsonEqual(v, v2) {
			t.Fatalf("round-trip mismatch\ninput:      %q\nre-encoded: %q", s, out)
		}
	})
}
