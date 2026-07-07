package toon_test

import (
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

func TestMarshalNestedStruct(t *testing.T) {
	type prefs struct {
		Theme string `toon:"theme"`
		Lang  string `toon:"lang"`
	}
	type profile struct {
		ID    int      `toon:"id"`
		Prefs prefs    `toon:"prefs"`
		Tags  []string `toon:"tags"`
	}

	in := profile{ID: 7, Prefs: prefs{Theme: "dark", Lang: "en"}, Tags: []string{"a", "b"}}
	got, err := toon.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	want := "id: 7\nprefs:\n  theme: dark\n  lang: en\ntags[2]: a,b"
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMarshalMapRoundTrip(t *testing.T) {
	in := map[string]int{"one": 1, "two": 2, "three": 3}
	data, err := toon.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]int
	if err := toon.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out["one"] != 1 || out["two"] != 2 || out["three"] != 3 {
		t.Fatalf("unexpected map: %v", out)
	}
}

func TestUnmarshalNumericKinds(t *testing.T) {
	input := []byte("u: 42\nf: 3.5\ni: -7")
	var out struct {
		U uint32  `toon:"u"`
		F float64 `toon:"f"`
		I int     `toon:"i"`
	}
	if err := toon.Unmarshal(input, &out); err != nil {
		t.Fatal(err)
	}
	if out.U != 42 || out.F != 3.5 || out.I != -7 {
		t.Fatalf("unexpected: %+v", out)
	}
}

// TestUnmarshalQuotedNumber exercises the string -> number conversion paths
// (asInt64/asFloat64 string branches) via quoted TOON values.
func TestUnmarshalQuotedNumber(t *testing.T) {
	input := []byte("f: \"3.5\"\ni: \"7\"")
	var out struct {
		F float64 `toon:"f"`
		I int     `toon:"i"`
	}
	if err := toon.Unmarshal(input, &out); err != nil {
		t.Fatal(err)
	}
	if out.F != 3.5 || out.I != 7 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestUnmarshalPointerField(t *testing.T) {
	input := []byte("name: Ada\ncount: 5")
	var out struct {
		Name  *string `toon:"name"`
		Count *int    `toon:"count"`
	}
	if err := toon.Unmarshal(input, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name == nil || *out.Name != "Ada" || out.Count == nil || *out.Count != 5 {
		t.Fatalf("unexpected: name=%v count=%v", out.Name, out.Count)
	}
}

func TestUnmarshalFixedArray(t *testing.T) {
	input := []byte("vals[3]: 1,2,3")
	var out struct {
		Vals [3]int `toon:"vals"`
	}
	if err := toon.Unmarshal(input, &out); err != nil {
		t.Fatal(err)
	}
	if out.Vals != [3]int{1, 2, 3} {
		t.Fatalf("unexpected: %v", out.Vals)
	}
}

func TestUnmarshalTypeMismatch(t *testing.T) {
	input := []byte("count: notanumber")
	var out struct {
		Count int `toon:"count"`
	}
	if err := toon.Unmarshal(input, &out); err == nil {
		t.Fatal("expected error decoding non-numeric string into int field")
	}
}
