package toon_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

type fixtureFile struct {
	Tests []fixtureCase `json:"tests"`
}

type fixtureCase struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected string          `json:"expected"`
	Options  fixtureOptions  `json:"options"`
}

type fixtureOptions struct {
	Indent        int    `json:"indent"`
	Delimiter     string `json:"delimiter"`
	LengthMarkers bool   `json:"lengthMarkers"`
}

func TestSpecEncodeFixtures(t *testing.T) {
	dir := filepath.Join("..", "testdata", "fixtures", "encode")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if entry.Name() == "key-folding.json" {
			continue // added when key folding lands
		}
		t.Run(entry.Name(), func(t *testing.T) {
			runFixtureFile(t, filepath.Join(dir, entry.Name()))
		})
	}
}

func runFixtureFile(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var file fixtureFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	for _, tc := range file.Tests {
		t.Run(tc.Name, func(t *testing.T) {
			value, err := toon.ParseJSON(tc.Input)
			if err != nil {
				t.Fatalf("parse input: %v", err)
			}

			opts := toon.EncodeOptions{
				Indent:        tc.Options.Indent,
				LengthMarkers: tc.Options.LengthMarkers,
			}
			if tc.Options.Delimiter != "" {
				opts.Delimiter = parseDelimiter(tc.Options.Delimiter)
			}

			got, err := toon.Marshal(value, opts)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if string(got) != tc.Expected {
				t.Fatalf("output mismatch\n--- got\n%s\n--- want\n%s", string(got), tc.Expected)
			}
		})
	}
}

func parseDelimiter(s string) toon.Delimiter {
	switch s {
	case "tab", "\t":
		return toon.DelimiterTab
	case "pipe", "|":
		return toon.DelimiterPipe
	default:
		return toon.DelimiterComma
	}
}

func TestEncodeJSONExample(t *testing.T) {
	in := []byte(`{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}`)
	want := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user"

	got, err := toon.EncodeJSON(in)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("got %q want %q", string(got), want)
	}
}
