# 04 — the encoder

`Marshal` walks an `Object` tree and writes into a `bytes.Buffer`.

## objects

Nested objects are indentation-based:

```text
user:
  id: 1
  name: Ada
```

Each field becomes either `key: value` on one line or `key:` followed by indented children.

## arrays

Three common shapes:

1. **Inline primitives** — `tags[3]: go, rust, zig`
2. **Tabular objects** — uniform rows share one header:
   ```text
   users[2]{id,name}:
     1,Ada
     2,Bob
   ```
3. **List items** — `-` prefixed lines for mixed/nested rows

`tabularFields` checks that every row is an object with the same primitive fields before we pick the tabular layout.

Tabular rows look up each column with a short loop over the row's fields — no per-row map allocation.

## indent cache

`writeIndent` caches `"  "`, `"    "`, … so we are not calling `strings.Repeat` on every line.

Next: read JSON into `Object` without scrambling key order.
