# go-toon

[![Go Reference](https://pkg.go.dev/badge/github.com/abhishektakale/go-toon/toon.svg)](https://pkg.go.dev/github.com/abhishektakale/go-toon/toon)

Go library and CLI for JSON → [**TOON**](https://toonformat.dev/) (Token-Oriented Object Notation) — a compact, LLM-friendly encoding of the JSON data model.

Repository: [github.com/abhishektakale/go-toon](https://github.com/abhishektakale/go-toon)

**Learning repo:** the library was built commit-by-commit; walk through it in [blog/](blog/).

## Install

**Library**

```bash
go get github.com/abhishektakale/go-toon/toon
```

**CLI**

```bash
go install github.com/abhishektakale/go-toon/cmd/go-toon@latest
```

## Quick start (library)

```go
package main

import (
	"fmt"
	"log"

	"github.com/abhishektakale/go-toon/toon"
)

func main() {
	jsonData := []byte(`{
		"users": [
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"}
		]
	}`)

	out, err := toon.EncodeJSON(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
```

Output:

```text
users[2]{id,name,role}:
  1,Alice,admin
  2,Bob,user
```

## API

| Function | Description |
|----------|-------------|
| `EncodeJSON(data []byte, opts ...EncodeOptions) ([]byte, error)` | Convert JSON bytes to TOON |
| `EncodeJSONFast(data []byte, opts ...EncodeOptions) ([]byte, error)` | Fast JSON parse (unordered keys) → TOON |
| `ParseJSON(data []byte) (any, error)` | Parse JSON with preserved key order |
| `ParseJSONFast(data []byte) (any, error)` | Fast JSON parse via `encoding/json` |
| `Marshal(v any, opts ...EncodeOptions) ([]byte, error)` | Encode a value as TOON (structs/maps supported) |
| `Decode(data []byte, opts ...DecodeOptions) (any, error)` | Parse TOON into Go values |
| `DecodeToJSON(data []byte, opts ...DecodeOptions) ([]byte, error)` | Convert TOON to compact JSON |
| `Unmarshal(data []byte, v any, opts ...DecodeOptions) error` | Decode TOON into a struct or map |

### Encode options

```go
toon.Marshal(value, toon.EncodeOptions{
	Indent:        2,                   // spaces per level (default: 2)
	Delimiter:     toon.DelimiterComma, // comma, tab, or pipe
	LengthMarkers: false,               // emit [#N] array headers
	KeyFolding:    toon.KeyFoldingSafe, // collapse single-key chains (§13.4)
	FlattenDepth:  toon.Int(2),         // optional max fold segments
})
```

### Decode options

```go
toon.Decode(data, toon.DecodeOptions{
	Indent:      2,
	Strict:      true,
	ExpandPaths: "safe", // split dotted keys back into nested objects
})
```

### Struct tags

```go
type User struct {
	ID   int    `toon:"id"`
	Name string `json:"name"`
}

out, _ := toon.Marshal(users)
var decoded Users
_ = toon.Unmarshal(toonData, &decoded)
```

## CLI

**Encode** (default):

```bash
go-toon encode < input.json > output.toon
go-toon < input.json > output.toon   # backward compatible
```

**Decode**:

```bash
go-toon decode < input.toon > output.json
go-toon decode -pretty < input.toon  # indented JSON
```

| Encode flag | Default | Description |
|-------------|---------|-------------|
| `-indent` | `2` | Spaces per indentation level |
| `-delimiter` | `comma` | `comma`, `tab`, or `pipe` |
| `-length-markers` | `false` | Emit `#` length markers |
| `-key-folding` | `off` | `off` or `safe` |
| `-fast` | `false` | Fast JSON parse (unordered keys) |

| Decode flag | Default | Description |
|-------------|---------|-------------|
| `-indent` | `2` | Expected spaces per level |
| `-delimiter` | `comma` | `comma`, `tab`, or `pipe` |
| `-strict` / `-no-strict` | `true` | Strict validation |
| `-expand-paths` | `off` | `off` or `safe` |
| `-pretty` | `false` | Pretty-print JSON output |

Build locally:

```bash
go build -o go-toon ./cmd/go-toon
```

## Examples

Runnable reference code lives under `examples/`:

```bash
go run ./examples/basic
```

See [examples/basic/main.go](examples/basic/main.go) for a minimal library integration.

## Project layout

```text
go-toon/
├── toon/                 # importable library (github.com/abhishektakale/go-toon/toon)
│   ├── doc.go            # package docs
│   ├── encode.go         # TOON encoder
│   ├── decode.go         # TOON decoder
│   ├── fold.go           # key folding (§13.4)
│   ├── reflect.go        # struct Marshal/Unmarshal
│   ├── json.go           # ordered + fast JSON parsing
│   ├── quote.go          # string/key quoting
│   ├── types.go          # public types and options
│   └── encode_bench_test.go  # performance benchmarks
├── blog/                 # step-by-step build notes
├── cmd/go-toon/          # CLI binary (encode + decode subcommands)
├── examples/basic/       # minimal library usage
├── .github/workflows/    # CI (go test on push/PR)
└── testdata/fixtures/    # official TOON spec encode + decode tests
```

## Conformance

The library is validated against the [official TOON specification fixtures](https://github.com/toon-format/spec/tree/main/tests/fixtures) for both encode and decode.

```bash
go test ./...
```

CI runs the same test suite on Go 1.22–1.26 via GitHub Actions (`.github/workflows/ci.yml`).

## Benchmarks

Run encoder benchmarks (throughput and allocations):

```bash
go test ./toon -bench=Benchmark -benchmem -run=^$
```

Sample results on `Intel Core Ultra 9 285K` / Windows / Go 1.26 (`go test ./toon -bench=Benchmark -benchmem -run=^$ -count=3`):

| Benchmark | Throughput | ns/op | B/op | allocs/op |
|-----------|------------|-------|------|-----------|
| `EncodeJSON_Small` | 18 MB/s | 3,314 | 2,840 | 77 |
| `EncodeJSON_Nested` | 27 MB/s | 7,350 | 5,632 | 196 |
| `EncodeJSON_Mixed` | 24 MB/s | 4,084 | 3,472 | 94 |
| `EncodeJSON_Tabular10` | 31 MB/s | 22,677 | 16,216 | 602 |
| `EncodeJSON_Tabular100` | 37 MB/s | 195,888 | 131,043 | 5,830 |
| `EncodeJSON_Tabular1000` | 38 MB/s | 1,948,933 | 1,324,805 | 58,940 |
| `ParseJSON_Tabular100` | 41 MB/s | 175,598 | 122,387 | 5,814 |
| `ParseJSONFast_Tabular100` | 102 MB/s | 70,400 | 82,345 | 1,620 |
| `EncodeJSONFast_Tabular100` | 82 MB/s | 87,910 | 90,994 | 1,636 |
| `Marshal_Tabular100` (encode only) | — | 15,556 | 8,654 | 16 |
| `Decode_Small` | 78 MB/s | 630 | 872 | 17 |
| `Decode_Tabular100` | 85 MB/s | 46,133 | 66,142 | 1,322 |
| `Decode_Tabular1000` | 83 MB/s | 502,097 | 660,294 | 13,047 |
| `DecodeToJSON_Tabular100` | 36 MB/s | 109,795 | 103,625 | 2,226 |
| `JSONMarshal_Tabular100` (baseline) | 164 MB/s | 43,878 | 37,109 | 904 |

For uniform tabular data (100 users), TOON output is ~**46% smaller** than compact JSON (3,907 vs 7,187 bytes) — the main win for LLM prompts. Encode-only (`Marshal` after parse) reuses the parsed tree and allocates far less per operation than the full `EncodeJSON` pipeline. When key order does not matter, `EncodeJSONFast` is roughly **2× faster** than ordered `EncodeJSON` on tabular data. Decoding tabular TOON runs at ~**85 MB/s** (`Decode`); `DecodeToJSON` adds the `encoding/json` marshal cost on top. Recent encoder optimizations (zero-alloc tabular rows, byte-scan quoting, indent cache) cut `Marshal` allocations from 115 to **16** per encode on tabular data.

Numbers vary by CPU and Go version; use the command above on your machine for local measurements.

## Comparison with other Go TOON libraries

Other Go implementations include [toon-format/toon-go](https://github.com/toon-format/toon-go) (community reference), [alpkeskin/gotoon](https://github.com/alpkeskin/gotoon), and [tnfssc/goon](https://github.com/tnfssc/goon). This library is optimized for the **JSON → TOON** path used in LLM pipelines.

| | **go-toon** | toon-go | alpkeskin/gotoon | goon |
|---|:---:|:---:|:---:|:---:|
| Zero dependencies | ✅ stdlib only | internal packages | standalone | standalone |
| `EncodeJSON` one-shot API | ✅ | ❌ (unmarshal + marshal) | ❌ | ❌ |
| Preserves JSON key order | ✅ | ❌ (`map` via `json.Unmarshal`) | ❌ | ❌ |
| Official spec encode fixtures | ✅ in-repo | separate | varies | varies |
| Official spec decode fixtures | ✅ in-repo | separate | varies | varies |
| Encode + decode | both | both | encode-focused | both |

### JSON → TOON throughput (100-row tabular users)

Local comparison on Intel Core Ultra 9 285K, Go 1.26 (`-count=3`). Peer libraries were measured with their typical `json.Unmarshal` + encode path; figures are indicative — run your own checks before drawing conclusions.

| Library | ns/op | allocs/op | vs go-toon (ordered) |
|---------|------:|----------:|----------------------|
| **go-toon** (ordered) | 167k | 5,830 | — |
| **go-toon** (`EncodeJSONFast`) | **96k** | **1,636** | **~1.7× faster** |
| toon-go | 115k | 3,422 | ~1.4× faster (ordered path) |
| alpkeskin/gotoon | 200k | 3,233 | go-toon faster |
| goon | 231k | 3,722 | go-toon faster |

With the fast JSON parse path (`EncodeJSONFast` / CLI `-fast`), go-toon is **competitive with or slightly ahead of toon-go** on full-pipeline tabular encoding when key order is not required. The ordered path improved ~**19%** after encoder optimizations; remaining gap vs toon-go is mostly ordered JSON parsing overhead.

### Encode-only (`Marshal` after parse)

When JSON is parsed once and TOON output is generated many times (caching, batch prompts), **go-toon is significantly leaner**:

| Library | ns/op | allocs/op |
|---------|------:|----------:|
| **go-toon** | **17k** | **16** |
| toon-go | 79k | 2,014 |

That is roughly **4.5× faster** with **~125× fewer allocations** per encode on tabular data.

### Where go-toon fits best

- **LLM prompt pipelines** — `EncodeJSON` converts API/DB JSON directly to TOON with correct key order; use `EncodeJSONFast` or `-fast` when order does not matter
- **Reference implementation** — spec fixtures, benchmarks, and examples in one repo
- **Low overhead hot loops** — parse once, call `Marshal` repeatedly with minimal allocation
- **Minimal dependencies** — no third-party modules in the core library

Reproduce the internal numbers with `go test ./toon -bench=Benchmark -benchmem -run=^$`.

## What is TOON?

TOON reduces token usage for structured LLM prompts by:

- Using indentation instead of braces
- Declaring array lengths and field names once for uniform object arrays
- Quoting strings only when necessary

See the [TOON specification](https://github.com/toon-format/spec/blob/main/SPEC.md) for details.

## License

MIT — see [LICENSE](LICENSE).
