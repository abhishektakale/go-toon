# 01 — why TOON and how this repo is laid out

TOON (Token-Oriented Object Notation) is still JSON underneath — same data model, fewer tokens when you feed it to an LLM.

A tabular JSON blob like:

```json
{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}
```

becomes something like:

```text
users[2]{id,name}:
  1,Alice
  2,Bob
```

The header declares the array length and field names once. Rows are just comma-separated values. That is the main win for prompt-sized payloads.

## what we are building

- `toon/` — importable package (`github.com/abhishektakale/go-toon/toon`)
- `cmd/go-toon/` — stdin/stdout CLI
- `testdata/fixtures/` — official spec tests from [toon-format/spec](https://github.com/toon-format/spec)
- `blog/` — these notes

No third-party deps in the core library. Stdlib only.

## design choices upfront

1. **Preserve JSON key order** when parsing. Go's `json.Unmarshal` into a map shuffles keys; TOON output should match the input order when it matters.
2. **Encoder-first** — get JSON → TOON solid before worrying about every decode edge case (we will get there).
3. **Small commits** — each step should compile and be explainable on its own.

Next up: the `Object` type and encode options.
