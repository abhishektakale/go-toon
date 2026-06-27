package toon

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Marshal turns a parsed value (usually Object) into TOON bytes.
func Marshal(v any, opts ...EncodeOptions) ([]byte, error) {
	var o EncodeOptions
	if len(opts) > 0 {
		o = opts[0].withDefaults()
	} else {
		o = EncodeOptions{}.withDefaults()
	}
	enc := newEncoder(o)
	return enc.encode(v)
}

type encoder struct {
	opts        EncodeOptions
	buf         bytes.Buffer
	indentCache []string
}

func newEncoder(opts EncodeOptions) *encoder {
	return &encoder{opts: opts.withDefaults()}
}

func (e *encoder) encode(v any) ([]byte, error) {
	switch val := v.(type) {
	case Object:
		if len(val) == 0 {
			return nil, nil
		}
		e.encodeObject(val, 0)
	case []any:
		e.encodeRootArray(val)
	default:
		e.writePrimitive(val, e.opts.Delimiter)
	}
	out := e.buf.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

func (e *encoder) writeIndent(depth int) {
	if depth <= 0 {
		return
	}
	idx := depth - 1
	for len(e.indentCache) <= idx {
		spaces := strings.Repeat(" ", (len(e.indentCache)+1)*e.opts.Indent)
		e.indentCache = append(e.indentCache, spaces)
	}
	e.buf.WriteString(e.indentCache[idx])
}

func (e *encoder) writePrimitive(v any, delim Delimiter) {
	switch val := v.(type) {
	case nil:
		e.buf.WriteString("null")
	case bool:
		if val {
			e.buf.WriteString("true")
		} else {
			e.buf.WriteString("false")
		}
	case string:
		writeQuotedPrimitive(&e.buf, val, delim)
	default:
		e.buf.WriteString(formatNumber(v))
	}
}

func (e *encoder) writeKey(key string) {
	writeKey(&e.buf, key)
}

func (e *encoder) delimiterSymbol(d Delimiter) string {
	switch d {
	case DelimiterTab:
		return "\t"
	case DelimiterPipe:
		return "|"
	default:
		return ""
	}
}

func (e *encoder) writeBracketSegment(length int, delim Delimiter) {
	e.buf.WriteByte('[')
	if e.opts.LengthMarkers {
		e.buf.WriteByte('#')
	}
	e.buf.WriteString(strconv.Itoa(length))
	if sym := e.delimiterSymbol(delim); sym != "" {
		e.buf.WriteString(sym)
	}
	e.buf.WriteByte(']')
}

func (e *encoder) writeFieldsSegment(fields []string, delim Delimiter) {
	if len(fields) == 0 {
		return
	}
	e.buf.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			e.buf.WriteRune(rune(delim))
		}
		e.writeKey(f)
	}
	e.buf.WriteByte('}')
}

func (e *encoder) writeArrayHeader(key string, length int, fields []string, delim Delimiter, hasKey bool) {
	if hasKey {
		e.writeKey(key)
	}
	e.writeBracketSegment(length, delim)
	e.writeFieldsSegment(fields, delim)
	e.buf.WriteByte(':')
}

func (e *encoder) writeColonAndInline(values []any, delim Delimiter) {
	e.buf.WriteByte(':')
	if len(values) > 0 {
		e.buf.WriteByte(' ')
		e.writeInlinePrimitives(values, delim)
	}
}

func (e *encoder) writeInlinePrimitives(values []any, delim Delimiter) {
	for i, v := range values {
		if i > 0 {
			e.buf.WriteRune(rune(delim))
		}
		e.writePrimitive(v, delim)
	}
}

func (e *encoder) encodeObject(obj Object, depth int) {
	for _, field := range obj {
		e.encodeField(field, depth, false)
	}
}

func (e *encoder) encodeField(field Field, depth int, onHyphenLine bool) {
	switch val := field.Value.(type) {
	case Object:
		if onHyphenLine {
			e.writeKey(field.Key)
			e.buf.WriteByte(':')
			e.buf.WriteByte('\n')
			if len(val) > 0 {
				e.encodeObject(val, depth+1)
			}
			return
		}
		e.writeIndent(depth)
		e.writeKey(field.Key)
		e.buf.WriteByte(':')
		e.buf.WriteByte('\n')
		if len(val) > 0 {
			e.encodeObject(val, depth+1)
		}
	case []any:
		e.encodeArray(field.Key, val, depth, onHyphenLine, e.opts.Delimiter, true)
	default:
		if onHyphenLine {
			e.writeKey(field.Key)
			e.buf.WriteString(": ")
			e.writePrimitive(val, e.opts.Delimiter)
			e.buf.WriteByte('\n')
			return
		}
		e.writeIndent(depth)
		e.writeKey(field.Key)
		e.buf.WriteString(": ")
		e.writePrimitive(val, e.opts.Delimiter)
		e.buf.WriteByte('\n')
	}
}

func (e *encoder) encodeRootArray(arr []any) {
	if len(arr) == 0 {
		e.buf.WriteString("[]")
		return
	}
	e.encodeArray("", arr, 0, false, e.opts.Delimiter, false)
}

func (e *encoder) encodeArray(key string, arr []any, depth int, onHyphenLine bool, docDelim Delimiter, hasKey bool) {
	if len(arr) == 0 {
		if !onHyphenLine {
			e.writeIndent(depth)
		}
		if hasKey {
			e.writeKey(key)
			e.buf.WriteString(": []")
			e.buf.WriteByte('\n')
		}
		return
	}

	if allPrimitive(arr) {
		if !onHyphenLine {
			e.writeIndent(depth)
		}
		if hasKey {
			e.writeKey(key)
		}
		e.writeBracketSegment(len(arr), docDelim)
		e.writeColonAndInline(arr, docDelim)
		e.buf.WriteByte('\n')
		return
	}

	if allPrimitiveArrays(arr) {
		if !onHyphenLine {
			e.writeIndent(depth)
		}
		if hasKey {
			e.writeKey(key)
		}
		e.writeBracketSegment(len(arr), docDelim)
		e.buf.WriteByte(':')
		e.buf.WriteByte('\n')
		for _, item := range arr {
			inner := item.([]any)
			e.writeIndent(depth + 1)
			e.buf.WriteString("- ")
			e.writeBracketSegment(len(inner), docDelim)
			e.writeColonAndInline(inner, docDelim)
			e.buf.WriteByte('\n')
		}
		return
	}

	if fields, ok := tabularFields(arr); ok {
		if !onHyphenLine {
			e.writeIndent(depth)
		}
		e.writeArrayHeader(key, len(arr), fields, docDelim, hasKey)
		e.buf.WriteByte('\n')
		for _, item := range arr {
			e.writeTabularRow(item.(Object), fields, depth+1, docDelim)
		}
		return
	}

	if !onHyphenLine {
		e.writeIndent(depth)
	}
	if hasKey {
		e.writeKey(key)
	}
	e.writeBracketSegment(len(arr), docDelim)
	e.buf.WriteByte(':')
	e.buf.WriteByte('\n')
	for _, item := range arr {
		e.encodeListItem(item, depth+1, docDelim)
	}
}

func (e *encoder) writeTabularRow(obj Object, fields []string, depth int, delim Delimiter) {
	e.writeIndent(depth)
	for i, name := range fields {
		if i > 0 {
			e.buf.WriteRune(rune(delim))
		}
		var val any
		for j := range obj {
			if obj[j].Key == name {
				val = obj[j].Value
				break
			}
		}
		e.writePrimitive(val, delim)
	}
	e.buf.WriteByte('\n')
}

func (e *encoder) encodeListItem(item any, depth int, docDelim Delimiter) {
	switch val := item.(type) {
	case Object:
		e.encodeObjectListItem(val, depth, docDelim)
	case []any:
		if allPrimitive(val) {
			e.writeIndent(depth)
			e.buf.WriteString("- ")
			e.writeBracketSegment(len(val), docDelim)
			e.writeColonAndInline(val, docDelim)
			e.buf.WriteByte('\n')
			return
		}
		e.writeIndent(depth)
		e.buf.WriteString("- ")
		e.writeBracketSegment(len(val), docDelim)
		e.buf.WriteByte(':')
		e.buf.WriteByte('\n')
		for _, inner := range val {
			e.encodeListItem(inner, depth+1, docDelim)
		}
	default:
		e.writeIndent(depth)
		e.buf.WriteString("- ")
		e.writePrimitive(val, docDelim)
		e.buf.WriteByte('\n')
	}
}

func (e *encoder) encodeObjectListItem(obj Object, depth int, docDelim Delimiter) {
	if len(obj) == 0 {
		e.writeIndent(depth)
		e.buf.WriteString("-")
		e.buf.WriteByte('\n')
		return
	}

	first := obj[0]
	if arr, ok := first.Value.([]any); ok {
		if fields, ok := tabularFields(arr); ok {
			e.writeIndent(depth)
			e.buf.WriteString("- ")
			e.writeArrayHeader(first.Key, len(arr), fields, docDelim, true)
			e.buf.WriteByte('\n')
			for _, row := range arr {
				e.writeTabularRow(row.(Object), fields, depth+2, docDelim)
			}
			for _, field := range obj[1:] {
				e.encodeField(field, depth+1, false)
			}
			return
		}
	}

	e.writeIndent(depth)
	e.buf.WriteString("- ")
	e.encodeField(first, depth+1, true)
	for _, field := range obj[1:] {
		e.encodeField(field, depth+1, false)
	}
}

func allPrimitive(arr []any) bool {
	for _, v := range arr {
		if !isPrimitive(v) {
			return false
		}
	}
	return true
}

func allPrimitiveArrays(arr []any) bool {
	for _, v := range arr {
		inner, ok := v.([]any)
		if !ok || !allPrimitive(inner) {
			return false
		}
	}
	return true
}

func isPrimitive(v any) bool {
	switch v.(type) {
	case nil, bool, string, json.Number, int, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

func tabularFields(arr []any) ([]string, bool) {
	if len(arr) == 0 {
		return nil, false
	}

	var fields []string
	keySet := map[string]struct{}{}

	for i, item := range arr {
		obj, ok := item.(Object)
		if !ok || len(obj) == 0 {
			return nil, false
		}
		if i == 0 {
			fields = make([]string, len(obj))
			for j, f := range obj {
				if !isPrimitive(f.Value) {
					return nil, false
				}
				fields[j] = f.Key
				keySet[f.Key] = struct{}{}
			}
			continue
		}
		if len(obj) != len(keySet) {
			return nil, false
		}
		seen := map[string]struct{}{}
		for _, f := range obj {
			if !isPrimitive(f.Value) {
				return nil, false
			}
			if _, ok := keySet[f.Key]; !ok {
				return nil, false
			}
			seen[f.Key] = struct{}{}
		}
		if len(seen) != len(keySet) {
			return nil, false
		}
	}
	return fields, true
}
