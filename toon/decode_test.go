package toon_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

type decodeFixtureFile struct {
	Tests []decodeFixtureCase `json:"tests"`
}

type decodeFixtureCase struct {
	Name        string          `json:"name"`
	Input       string          `json:"input"`
	Expected    json.RawMessage `json:"expected"`
	ShouldError bool            `json:"shouldError"`
	Options     decodeOptions   `json:"options"`
}

type decodeOptions struct {
	Indent      int    `json:"indent"`
	Strict      *bool  `json:"strict"`
	ExpandPaths string `json:"expandPaths"`
	Delimiter   string `json:"delimiter"`
}

func TestSpecDecodeFixtures(t *testing.T) {
	dir := filepath.Join("..", "testdata", "fixtures", "decode")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			runDecodeFixtureFile(t, filepath.Join(dir, entry.Name()))
		})
	}
}

func runDecodeFixtureFile(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var file decodeFixtureFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, tc := range file.Tests {
		t.Run(tc.Name, func(t *testing.T) {
			opts := toon.DecodeOptions{ExpandPaths: tc.Options.ExpandPaths}
			if tc.Options.Indent > 0 {
				opts.Indent = tc.Options.Indent
			}
			if tc.Options.Strict != nil {
				opts.Strict = *tc.Options.Strict
			} else {
				opts.Strict = true
			}
			if tc.Options.Delimiter != "" {
				opts.Delimiter = parseDelimiter(tc.Options.Delimiter)
			}

			got, err := toon.Decode([]byte(tc.Input), opts)
			if tc.ShouldError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("decode: %v", err)
			}

			var want any
			if err := json.Unmarshal(tc.Expected, &want); err != nil {
				t.Fatalf("parse expected: %v", err)
			}
			if !jsonEqual(got, want) {
				gj, _ := json.Marshal(got)
				wj, _ := json.Marshal(want)
				t.Fatalf("output mismatch\n--- got\n%s\n--- want\n%s", string(gj), string(wj))
			}
		})
	}
}

func jsonEqual(a, b any) bool {
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	var va, vb any
	if json.Unmarshal(aj, &va) != nil || json.Unmarshal(bj, &vb) != nil {
		return false
	}
	return reflectDeepEqual(va, vb)
}

func reflectDeepEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}

func TestRoundTrip_TabularUsers(t *testing.T) {
	in := []byte(`{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}`)
	toonBytes, err := toon.EncodeJSON(in)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := toon.Decode(toonBytes)
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	json.Unmarshal(out, &got)
	json.Unmarshal(in, &want)
	if !jsonEqual(got, want) {
		t.Fatalf("round trip mismatch: %s vs %s", out, in)
	}
}
