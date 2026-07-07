package toon_test

import (
	"testing"

	"github.com/abhishektakale/go-toon/toon"
)

func benchUsersTOON(b *testing.B, rows int) []byte {
	b.Helper()
	data, err := toon.EncodeJSON(benchUsersJSON(rows))
	if err != nil {
		b.Fatal(err)
	}
	return data
}

func BenchmarkDecode_Small(b *testing.B) {
	data, err := toon.EncodeJSON(benchSmallJSON)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.Decode(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Tabular100(b *testing.B) {
	data := benchUsersTOON(b, 100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.Decode(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Tabular1000(b *testing.B) {
	data := benchUsersTOON(b, 1000)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.Decode(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeToJSON_Tabular100(b *testing.B) {
	data := benchUsersTOON(b, 100)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		if _, err := toon.DecodeToJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}
