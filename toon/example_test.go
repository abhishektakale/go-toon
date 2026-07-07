package toon_test

import (
	"fmt"

	"github.com/abhishektakale/go-toon/toon"
)

func ExampleEncodeJSON() {
	data := []byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`)
	out, err := toon.EncodeJSON(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	// Output:
	// users[2]{id,name}:
	//   1,Alice
	//   2,Bob
}

func ExampleDecodeToJSON() {
	data := []byte("users[2]{id,name}:\n  1,Alice\n  2,Bob")
	out, err := toon.DecodeToJSON(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	// Output: {"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}
}
