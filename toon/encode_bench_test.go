package toon_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

var (
	benchSmallJSON = []byte(`{"id":123,"name":"Ada","active":true,"tags":["admin","ops"]}`)

	benchNestedJSON = []byte(`{
		"user": {
			"id": 123,
			"name": "Ada",
			"prefs": {"theme": "dark", "lang": "en"}
		},
		"items": [
			{"sku": "A1", "qty": 2, "price": 9.99},
			{"sku": "B2", "qty": 1, "price": 14.5}
		]
	}`)

	benchMixedJSON = []byte(`{
		"items": [
			1,
			{"a": 1},
			"text",
			{"id": 2, "name": "Second", "extra": true}
		]
	}`)
)

func benchUsersJSON(rows int) []byte {
	var b bytes.Buffer
	b.WriteString(`{"users":[`)
	for i := 0; i < rows; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"name":"User %d","email":"user%d@example.com","role":"member"}`, i+1, i+1, i+1)
	}
	b.WriteString(`]}`)
	return b.Bytes()
}

func BenchmarkEncodeJSON_Small(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchSmallJSON)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(benchSmallJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSON_Nested(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchNestedJSON)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(benchNestedJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSON_Mixed(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(benchMixedJSON)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(benchMixedJSON); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSON_Tabular10(b *testing.B) {
	data := benchUsersJSON(10)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSON_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSON_Tabular1000(b *testing.B) {
	data := benchUsersJSON(1000)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseJSON_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.ParseJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseJSONFast_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.ParseJSONFast(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSONFast_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.EncodeJSONFast(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	value, err := toon.ParseJSON(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := toon.Marshal(value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshal_Tabular100(b *testing.B) {
	data := benchUsersJSON(100)
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(payload); err != nil {
			b.Fatal(err)
		}
	}
}

func TestOutputSizeReduction_Tabular100(t *testing.T) {
	data := benchUsersJSON(100)
	jsonCompact, err := json.Marshal(json.RawMessage(data))
	if err != nil {
		t.Fatal(err)
	}
	toonOut, err := toon.EncodeJSON(data)
	if err != nil {
		t.Fatal(err)
	}

	reduction := 100 * float64(len(jsonCompact)-len(toonOut)) / float64(len(jsonCompact))
	t.Logf("json=%d bytes toon=%d bytes reduction=%.1f%%", len(jsonCompact), len(toonOut), reduction)

	if len(toonOut) >= len(jsonCompact) {
		t.Fatalf("expected TOON output to be smaller than JSON for tabular data")
	}
}
