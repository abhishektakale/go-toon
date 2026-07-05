// Basic example: convert JSON to TOON using the library API.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abhishektakale/go-toon/toon"
)

func main() {
	jsonData, err := os.ReadFile("input.json")
	if err != nil {
		// Fall back to inline sample when no file is provided.
		jsonData = []byte(`{
			"users": [
				{"id": 1, "name": "Alice", "role": "admin"},
				{"id": 2, "name": "Bob", "role": "user"}
			]
		}`)
	}

	out, err := toon.EncodeJSON(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
}
