# 06 — decoding TOON

Encoding was the fun part first because that is the LLM prompt path. Decoding mirrors it:

1. Split input into lines, track indentation depth
2. Parse headers like `users[2]{id,name}:` and inline `tags[3]: a,b,c`
3. Rebuild `map[string]any` / `[]any` trees

`DecodeOptions` covers strict mode, delimiter, and optional `expandPaths: safe` to undo dotted keys from folding.

`DecodeToJSON` is a thin wrapper around `encoding/json.Marshal`.

Tests use the official decode fixtures under `testdata/fixtures/decode/`.

Next: CLI and struct helpers.
