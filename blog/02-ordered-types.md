# 02 — ordered objects instead of maps

Go maps randomize iteration order. TOON cares about key order when mirroring JSON, so we keep objects as a slice of fields:

```go
type Field struct {
    Key   string
    Value any
}

type Object []Field
```

Values are plain `any` — nested `Object`, `[]any`, or primitives (`string`, `bool`, `json.Number`, etc.).

`EncodeOptions` holds the knobs we will need later:

- `Indent` — spaces per level (default 2)
- `Delimiter` — comma, tab, or pipe between inline values
- `LengthMarkers` — optional `#` prefix on array lengths

Nothing encodes yet. Next: when strings need quotes.
