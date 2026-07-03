package toon_test

import (
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

type user struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Role string `json:"role"`
}

func TestMarshalStruct(t *testing.T) {
	in := struct {
		Users []user `toon:"users"`
	}{
		Users: []user{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "user"},
		},
	}

	got, err := toon.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	want := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user"
	if string(got) != want {
		t.Fatalf("got %q want %q", string(got), want)
	}
}

func TestUnmarshalStruct(t *testing.T) {
	input := []byte("users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user")
	var out struct {
		Users []user `toon:"users"`
	}
	if err := toon.Unmarshal(input, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Users) != 2 || out.Users[0].Name != "Alice" || out.Users[1].Role != "user" {
		t.Fatalf("unexpected struct: %+v", out)
	}
}

func TestParseJSONFastProducesTOON(t *testing.T) {
	in := []byte(`{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}`)
	got, err := toon.EncodeJSONFast(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected output")
	}
}
