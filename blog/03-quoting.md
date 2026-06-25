# 03 — quoting strings and keys

TOON only quotes when it has to. A value needs quotes if it:

- is empty or has leading/trailing whitespace
- looks like `true`, `false`, `null`, or a number
- contains structural characters (`:`, `[`, `{`, `"`, `\`)
- starts with `-` (so it is not confused with a list marker)
- contains the active array delimiter

Keys use a slightly looser rule: unquoted if they match identifier-ish characters (letters, digits, `_`, `.`).

We scan bytes instead of compiling regexes — hot path for tabular data with lots of short strings.

Helper writers (`writeKey`, `writeQuotedPrimitive`) write straight into a `bytes.Buffer` via `io.Writer`.

Next: actually emit TOON lines.
