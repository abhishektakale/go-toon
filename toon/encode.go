package toon

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Marshal encodes a Go value as TOON.
func Marshal(v any, opts ...EncodeOptions) ([]byte, error) {
	normalized, err := normalizeForEncode(v)
	if err != nil {
		return nil, err
	}
	var o EncodeOptions
	if len(opts) > 0 {
		o = opts[0].withDefaults()
	} else {
		o = EncodeOptions{}.withDefaults()
	}
	enc := newEncoder(o)
	return enc.encode(normalized)
}

type encoder struct {
	opts         EncodeOptions
	buf          bytes.Buffer
	rootSiblings map[string]struct{}
	indentCache  []string
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
		e.rootSiblings = siblingKeySet(val)
		e.encodeObject(val, 0, "", true)
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

func (e *encoder) encodeObject(obj Object, depth int, pathPrefix string, allowFold bool) {
	fields := make([]encField, len(obj))
	for i, f := range obj {
		fields[i] = encField{field: f}
	}
	if allowFold {
		if maxSeg := maxFoldSegments(e.opts); maxSeg != 0 {
			limit := maxSeg
			if limit < 0 {
				limit = 1<<31 - 1
			}
			fields = foldObject(obj, siblingKeySet(obj), limit, pathPrefix, e.rootSiblings)
		}
	}
	for _, item := range fields {
		e.encodeField(item.field, depth, false, pathPrefix, allowFold && !item.noNestedFold)
	}
}

func (e *encoder) encodeField(field Field, depth int, onHyphenLine bool, pathPrefix string, allowFold bool) {
	childPrefix := joinPath(pathPrefix, field.Key)
	if m, ok := field.Value.(map[string]any); ok {
		field.Value = mapToObject(m)
	}
	switch val := field.Value.(type) {
	case Object:
		if onHyphenLine {
			e.writeKey(field.Key)
			e.buf.WriteByte(':')
			e.buf.WriteByte('\n')
			if len(val) > 0 {
				e.encodeObject(val, depth+1, childPrefix, allowFold)
			}
			return
		}
		e.writeIndent(depth)
		e.writeKey(field.Key)
		e.buf.WriteByte(':')
		e.buf.WriteByte('\n')
		if len(val) > 0 {
			e.encodeObject(val, depth+1, childPrefix, allowFold)
		}
	case []any:
		e.encodeArray(field.Key, val, depth, onHyphenLine, e.opts.Delimiter, true, childPrefix, allowFold)
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
	e.encodeArray("", arr, 0, false, e.opts.Delimiter, false, "", true)
}

func (e *encoder) encodeArray(key string, arr []any, depth int, onHyphenLine bool, docDelim Delimiter, hasKey bool, pathPrefix string, allowFold bool) {
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
			obj, _ := asObject(item)
			e.writeTabularRow(obj, fields, depth+1, docDelim)
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
		e.encodeListItem(item, depth+1, docDelim, pathPrefix, allowFold)
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

func (e *encoder) encodeListItem(item any, depth int, docDelim Delimiter, pathPrefix string, allowFold bool) {
	if m, ok := item.(map[string]any); ok {
		item = mapToObject(m)
	}
	switch val := item.(type) {
	case Object:
		e.encodeObjectListItem(val, depth, docDelim, pathPrefix, allowFold)
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
			e.encodeListItem(inner, depth+1, docDelim, pathPrefix, allowFold)
		}
	default:
		e.writeIndent(depth)
		e.buf.WriteString("- ")
		e.writePrimitive(val, docDelim)
		e.buf.WriteByte('\n')
	}
}

func (e *encoder) encodeObjectListItem(obj Object, depth int, docDelim Delimiter, pathPrefix string, allowFold bool) {
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
				obj, _ := asObject(row)
				e.writeTabularRow(obj, fields, depth+2, docDelim)
			}
			for _, field := range obj[1:] {
				e.encodeField(field, depth+1, false, pathPrefix, allowFold)
			}
			return
		}
	}

	e.writeIndent(depth)
	e.buf.WriteString("- ")
	e.encodeField(first, depth+1, true, pathPrefix, allowFold)
	for _, field := range obj[1:] {
		e.encodeField(field, depth+1, false, pathPrefix, allowFold)
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

// mapToObject converts the unordered map form produced by Decode into the
// ordered Object the encoder works with. Key order is unspecified.
func mapToObject(m map[string]any) Object {
	obj := make(Object, 0, len(m))
	for k, v := range m {
		obj = append(obj, Field{Key: k, Value: v})
	}
	return obj
}

// asObject accepts either object representation the encoder may receive: the
// ordered Object from ParseJSON/reflect, or the map[string]any from Decode.
func asObject(v any) (Object, bool) {
	switch val := v.(type) {
	case Object:
		return val, true
	case map[string]any:
		return mapToObject(val), true
	default:
		return nil, false
	}
}

func tabularFields(arr []any) ([]string, bool) {
	if len(arr) == 0 {
		return nil, false
	}

	var fields []string
	keySet := map[string]struct{}{}

	for i, item := range arr {
		obj, ok := asObject(item)
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
