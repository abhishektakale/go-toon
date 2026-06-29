# 05 — parsing JSON without losing key order

`encoding/json` is fine for many jobs, but unmarshaling into `map[string]any` loses key order.

We use `json.Decoder` in streaming mode:

1. `Token()` for `{`, `[`, keys, literals
2. Build `Object` as a slice of `Field` in encounter order
3. `UseNumber()` so integers do not become `float64`

`EncodeJSON` is the convenience wrapper: `ParseJSON` then `Marshal`.

## tests

Official encode fixtures live in `testdata/fixtures/encode/`. The test harness loads each JSON file and compares our output to the expected TOON string.

Next: decoding TOON back into Go values.
