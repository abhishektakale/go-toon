# 07 — structs, key folding, fast JSON

Three features that sit on top of the core encoder/decoder:

## Marshal / Unmarshal

`reflect` walks structs and maps into `Object` before encode. `toon` struct tags name fields (falls back to `json` tags). `Unmarshal` fills structs from decoded maps.

## Key folding

Optional `KeyFolding: safe` collapses chains like `{a:{b:{c:1}}}` into `a.b.c: 1` when segments are safe identifiers and nothing collides with sibling keys. Pair with `expandPaths: safe` on decode for round trips.

## ParseJSONFast

When key order does not matter, skip the streaming tokenizer and use `json.Decode` into `map[string]any`, then convert once to `Object`. `EncodeJSONFast` exposes it; the CLI has `-fast`.

Next: command-line tool and benchmarks.
