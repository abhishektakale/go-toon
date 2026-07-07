# 08 — shipping it

## CLI

`cmd/go-toon` exposes `encode` and `decode` subcommands reading stdin / writing stdout. Flags mirror library options (`-indent`, `-delimiter`, `-key-folding`, `-fast`, `-expand-paths`).

## benchmarks

`go test ./toon -bench=Benchmark -benchmem` tracks encoder cost. Cross-library numbers in the README come from one-off local runs against [toon-go](https://github.com/toon-format/toon-go), [alpkeskin/gotoon](https://github.com/alpkeskin/gotoon), and [goon](https://github.com/tnfssc/goon) — not checked into this repo.

Recent numbers on a Core Ultra 9 285K (Go 1.26):

- `Marshal_Tabular100` — ~18µs, 16 allocs
- `EncodeJSON_Tabular100` — ~200µs (parse still dominates)
- `EncodeJSONFast_Tabular100` — ~96µs when key order is not needed

## CI

`.github/workflows/ci.yml` runs `go test ./...` on push/PR across Go 1.22–1.26.

That is the full loop for v1. Further work (custom ordered JSON tokenizer, more round-trip tests) can build on the same commit-by-commit style.
