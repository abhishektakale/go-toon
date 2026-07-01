package toon

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// DecodeOptions configures TOON decoding.
type DecodeOptions struct {
	Indent        int
	Strict        bool
	Delimiter     Delimiter
	ExpandPaths   string // "off" or "safe"
}

func (o DecodeOptions) withDefaults() DecodeOptions {
	if o.Indent <= 0 {
		o.Indent = 2
	}
	if o.Delimiter == 0 {
		o.Delimiter = DelimiterComma
	}
	if o.ExpandPaths == "" {
		o.ExpandPaths = "off"
	}
	return o
}

// Decode parses a TOON document into Go values (map[string]any, []any, primitives).
func Decode(data []byte, opts ...DecodeOptions) (any, error) {
	var o DecodeOptions
	if len(opts) > 0 {
		o = opts[0].withDefaults()
	} else {
		o = DecodeOptions{Strict: true}.withDefaults()
	}
	p, err := newParser(string(data), o)
	if err != nil {
		return nil, err
	}
	return p.parseDocument()
}

// DecodeToJSON converts TOON bytes to compact JSON.
func DecodeToJSON(data []byte, opts ...DecodeOptions) ([]byte, error) {
	v, err := Decode(data, opts...)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

type parser struct {
	lines []parsedLine
	pos   int
	cfg   DecodeOptions
}

type parsedLine struct {
	number  int
	indent  int
	content string
	blank   bool
}

func newParser(input string, cfg DecodeOptions) (*parser, error) {
	rawLines := splitLines(input)
	lines := make([]parsedLine, 0, len(rawLines))
	for idx, raw := range rawLines {
		if raw == "" {
			lines = append(lines, parsedLine{number: idx + 1, blank: true})
			continue
		}
		indent, content, err := computeIndent(raw, cfg)
		if err != nil {
			return nil, errorWrap(idx+1, err)
		}
		lines = append(lines, parsedLine{
			number:  idx + 1,
			indent:  indent,
			content: content,
			blank:   strings.TrimSpace(content) == "",
		})
	}
	return &parser{lines: lines, cfg: cfg}, nil
}

func splitLines(input string) []string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	lines := strings.Split(input, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func computeIndent(line string, cfg DecodeOptions) (int, string, error) {
	indent := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			indent++
		case '\t':
			if cfg.Strict {
				return 0, "", errors.New("tabs are not allowed in indentation (strict mode)")
			}
			indent++
		default:
			content := line[i:]
			if cfg.Strict && indent%cfg.Indent != 0 {
				return 0, "", fmt.Errorf("indentation must be a multiple of %d spaces", cfg.Indent)
			}
			return indent / cfg.Indent, content, nil
		}
	}
	return 0, "", nil
}

func (p *parser) parseDocument() (any, error) {
	p.skipBlankLinesOutsideArrays()
	if p.pos >= len(p.lines) {
		return map[string]any{}, nil
	}

	nonBlank := p.countRemainingNonBlank()
	first := p.current()

	if nonBlank == 1 && strings.TrimSpace(first.content) == "[]" {
		p.pos++
		return []any{}, nil
	}

	header, ok, err := tryParseHeader(first.content, p.cfg.Strict)
	if err != nil {
		return nil, errorWrap(first.number, err)
	}

	if nonBlank == 1 && !ok && !isKeyValue(first.content) {
		token := strings.TrimSpace(first.content)
		value, err := decodePrimitiveToken(token)
		if err != nil {
			return nil, errorWrap(first.number, err)
		}
		p.pos++
		return value, nil
	}

	if ok && first.indent == 0 && !header.hasKey {
		p.pos++
		return p.parseArray(header, 0, false)
	}

	return p.parseObject(0)
}

func (p *parser) parseObject(depth int) (map[string]any, error) {
	result := make(map[string]any)
	for p.pos < len(p.lines) {
		line := p.current()
		if line.blank {
			p.pos++
			continue
		}
		if line.indent < depth {
			break
		}
		if line.indent > depth {
			return nil, errorAt(line.number, "unexpected indentation")
		}

		header, isHeader, err := tryParseHeader(line.content, p.cfg.Strict)
		if err != nil {
			if p.cfg.Strict {
				return nil, errorWrap(line.number, err)
			}
			isHeader = false
		}
		if isHeader {
			if !header.hasKey {
				return nil, errorAt(line.number, "arrays within objects must have a key")
			}
			p.pos++
			value, err := p.parseArray(header, depth, false)
			if err != nil {
				return nil, err
			}
			if err := p.setKey(result, header.key, value, header.keyQuoted); err != nil {
				return nil, errorWrap(line.number, err)
			}
			continue
		}

		key, rest, keyQuoted, err := p.splitKeyValue(line.content)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		p.pos++

		if rest == "" {
			nextValue, err := p.parseObject(depth + 1)
			if err != nil {
				return nil, err
			}
			if err := p.setKey(result, key, nextValue, keyQuoted); err != nil {
				return nil, err
			}
			continue
		}

		if rest == "[]" {
			if err := p.setKey(result, key, []any{}, keyQuoted); err != nil {
				return nil, err
			}
			continue
		}

		value, err := decodePrimitiveToken(rest)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		if err := p.setKey(result, key, value, keyQuoted); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (p *parser) setKey(obj map[string]any, key string, value any, keyQuoted bool) error {
	if p.cfg.ExpandPaths == "safe" && !keyQuoted && isExpandablePath(key) {
		return mergeExpanded(obj, key, value, p.cfg.Strict)
	}
	if p.cfg.Strict {
		if _, exists := obj[key]; exists {
			return errorAt(0, fmt.Sprintf("duplicate key %q", key))
		}
	}
	obj[key] = value
	return nil
}

func isExpandablePath(key string) bool {
	if key == "" || strings.Contains(key, `"`) {
		return false
	}
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if !isIdentifierSegment(part) {
			return false
		}
	}
	return true
}

func mergeExpanded(root map[string]any, path string, value any, strict bool) error {
	parts := strings.Split(path, ".")
	cur := root
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := cur[part]
		if !ok {
			nested := make(map[string]any)
			cur[part] = nested
			cur = nested
			continue
		}
		nested, ok := next.(map[string]any)
		if !ok {
			if strict {
				return errorAt(0, fmt.Sprintf("expansion conflict at %q", part))
			}
			nested = make(map[string]any)
			cur[part] = nested
		}
		cur = nested
	}
	last := parts[len(parts)-1]
	if strict {
		if existing, ok := cur[last]; ok {
			if _, isObj := existing.(map[string]any); isObj {
				return errorAt(0, fmt.Sprintf("expansion conflict at %q", last))
			}
			return errorAt(0, fmt.Sprintf("duplicate key %q", last))
		}
	}
	cur[last] = value
	return nil
}

func (p *parser) parseArray(header parsedHeader, depth int, fromHyphen bool) (any, error) {
	delimiter := rune(header.delimiter)
	rowDepth := depth + 1
	itemDepth := depth + 1
	if fromHyphen {
		rowDepth = depth + 2
		itemDepth = depth + 2
	}
	var values []any

	if len(header.inlineValues) > 0 {
		raw, err := SplitInlineValues(header.inlineValues, delimiter)
		if err != nil {
			return nil, errorWrap(p.lines[p.pos-1].number, err)
		}
		for _, token := range raw {
			value, err := decodePrimitiveToken(token)
			if err != nil {
				return nil, errorWrap(p.lines[p.pos-1].number, err)
			}
			values = append(values, value)
		}
		if p.cfg.Strict && len(values) != header.length {
			return nil, errorAtf(p.lines[p.pos-1].number, "inline array length mismatch; expected %d, got %d", header.length, len(values))
		}
		return values, nil
	}

	if len(header.fields) > 0 {
		rows := make([]any, 0, header.length)
		for p.pos < len(p.lines) {
			line := p.current()
			if line.blank {
				if p.cfg.Strict {
					if nextIndent, ok := p.nextNonBlankIndent(p.pos); !ok || nextIndent <= depth {
						break
					}
					return nil, errorAt(line.number, "blank line inside tabular array")
				}
				p.pos++
				continue
			}
			if line.indent <= depth {
				break
			}
			if line.indent < rowDepth {
				break
			}
			if line.indent != rowDepth {
				return nil, errorAt(line.number, "invalid indentation for tabular row")
			}
			trimmed := strings.TrimSpace(line.content)
			if indexOutsideQuotes(trimmed, ':') != -1 {
				break
			}
			p.pos++
			raw, err := SplitInlineValues(trimmed, delimiter)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if p.cfg.Strict && len(raw) != len(header.fields) {
				return nil, errorAt(line.number, "tabular row width mismatch")
			}
			row := make(map[string]any, len(header.fields))
			for idx, field := range header.fields {
				if idx >= len(raw) {
					break
				}
				value, err := decodePrimitiveToken(raw[idx])
				if err != nil {
					return nil, errorWrap(line.number, err)
				}
				row[field] = value
			}
			rows = append(rows, row)
			if p.cfg.Strict && len(rows) > header.length {
				return nil, errorAtf(line.number, "too many tabular rows (expected %d)", header.length)
			}
		}
		if p.cfg.Strict && len(rows) != header.length {
			lineNo := 0
			if p.pos > 0 {
				lineNo = p.lines[p.pos-1].number
			}
			return nil, errorAtf(lineNo, "tabular length mismatch; expected %d rows", header.length)
		}
		return rows, nil
	}

	values = make([]any, 0, header.length)
	for p.pos < len(p.lines) {
		line := p.current()
		if line.blank {
			if p.cfg.Strict {
				if nextIndent, ok := p.nextNonBlankIndent(p.pos); !ok || nextIndent <= depth {
					break
				}
				return nil, errorAt(line.number, "blank line inside list array")
			}
			p.pos++
			continue
		}
		if line.indent <= depth {
			break
		}
		if line.indent < itemDepth {
			break
		}
		if line.indent != itemDepth {
			return nil, errorAt(line.number, "invalid indentation for list item")
		}
		if !strings.HasPrefix(line.content, "-") {
			break
		}
		itemContent := strings.TrimSpace(line.content[1:])
		p.pos++
		if itemContent == "" {
			values = append(values, map[string]any{})
			continue
		}

		if strings.HasPrefix(itemContent, "[") {
			itemHeader, ok, err := tryParseHeader(itemContent, p.cfg.Strict)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if ok && !itemHeader.hasKey {
				itemValue, err := p.parseArray(itemHeader, depth+1, false)
				if err != nil {
					return nil, err
				}
				values = append(values, itemValue)
				continue
			}
		}

		if itemHeader, isHeader, err := tryParseHeader(itemContent, p.cfg.Strict); err != nil {
			return nil, errorWrap(line.number, err)
		} else if isHeader {
			if !itemHeader.hasKey {
				itemValue, err := p.parseArray(itemHeader, depth+1, false)
				if err != nil {
					return nil, err
				}
				values = append(values, itemValue)
				continue
			}
			arrayValue, err := p.parseArray(itemHeader, depth+1, true)
			if err != nil {
				return nil, err
			}
			obj := map[string]any{itemHeader.key: arrayValue}
			if err := p.collectObjectListSiblings(obj, depth); err != nil {
				return nil, err
			}
			values = append(values, obj)
			continue
		}

		if isKeyValue(itemContent) {
			key, rest, _, err := p.splitKeyValue(itemContent)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			if rest == "" {
				obj, err := p.parseObject(depth + 3)
				if err != nil {
					return nil, err
				}
				values = append(values, map[string]any{key: obj})
				continue
			}
			if rest == "[]" {
				obj := map[string]any{key: []any{}}
				if err := p.collectObjectListSiblings(obj, depth); err != nil {
					return nil, err
				}
				values = append(values, obj)
				continue
			}
			val, err := decodePrimitiveToken(rest)
			if err != nil {
				return nil, errorWrap(line.number, err)
			}
			obj := map[string]any{key: val}
			if err := p.collectObjectListSiblings(obj, depth); err != nil {
				return nil, err
			}
			values = append(values, obj)
			continue
		}

		value, err := decodePrimitiveToken(itemContent)
		if err != nil {
			return nil, errorWrap(line.number, err)
		}
		values = append(values, value)
	}

	if p.cfg.Strict && len(values) != header.length {
		lineNo := 0
		if p.pos > 0 && p.pos <= len(p.lines) {
			lineNo = p.lines[p.pos-1].number
		}
		return nil, errorAtf(lineNo, "list length mismatch; expected %d items", header.length)
	}
	return values, nil
}

func (p *parser) current() parsedLine {
	return p.lines[p.pos]
}

func (p *parser) skipBlankLinesOutsideArrays() {
	for p.pos < len(p.lines) {
		if !p.lines[p.pos].blank {
			break
		}
		p.pos++
	}
}

func (p *parser) countRemainingNonBlank() int {
	count := 0
	for _, line := range p.lines[p.pos:] {
		if !line.blank {
			count++
		}
	}
	return count
}

func (p *parser) nextNonBlankIndent(from int) (int, bool) {
	for i := from + 1; i < len(p.lines); i++ {
		if !p.lines[i].blank {
			return p.lines[i].indent, true
		}
	}
	return 0, false
}

func (p *parser) collectObjectListSiblings(obj map[string]any, depth int) error {
	for p.pos < len(p.lines) {
		next := p.current()
		if next.blank {
			if p.cfg.Strict {
				if ni, ok := p.nextNonBlankIndent(p.pos); !ok || ni <= depth+1 {
					break
				}
				return errorAt(next.number, "blank line inside object list item")
			}
			p.pos++
			continue
		}
		if next.indent <= depth+1 {
			break
		}
		if next.indent != depth+2 {
			return errorAt(next.number, "invalid indentation for object list sibling")
		}
		if header, isHeader, err := tryParseHeader(next.content, p.cfg.Strict); err != nil {
			return errorWrap(next.number, err)
		} else if isHeader {
			p.pos++
			value, err := p.parseArray(header, depth+1, false)
			if err != nil {
				return err
			}
			if header.key == "" {
				return errorAt(next.number, "arrays within objects must have a key")
			}
			obj[header.key] = value
			continue
		}
		key, rest, _, err := p.splitKeyValue(next.content)
		if err != nil {
			return errorWrap(next.number, err)
		}
		p.pos++
		if rest == "" {
			nested, err := p.parseObject(depth + 3)
			if err != nil {
				return err
			}
			if err := p.setObjField(obj, key, nested, next.number); err != nil {
				return err
			}
		} else if rest == "[]" {
			if err := p.setObjField(obj, key, []any{}, next.number); err != nil {
				return err
			}
		} else {
			value, err := decodePrimitiveToken(rest)
			if err != nil {
				return errorWrap(next.number, err)
			}
			if err := p.setObjField(obj, key, value, next.number); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *parser) setObjField(obj map[string]any, key string, value any, line int) error {
	if p.cfg.Strict {
		if _, exists := obj[key]; exists {
			return errorAt(line, fmt.Sprintf("duplicate key %q", key))
		}
	}
	obj[key] = value
	return nil
}

type parsedHeader struct {
	key          string
	hasKey       bool
	keyQuoted    bool
	length       int
	delimiter    Delimiter
	fields       []string
	inlineValues string
}

func tryParseHeader(content string, strict bool) (parsedHeader, bool, error) {
	colon := indexOutsideQuotes(content, ':')
	if colon == -1 {
		return parsedHeader{}, false, nil
	}
	left := strings.TrimSpace(content[:colon])
	right := strings.TrimSpace(content[colon+1:])
	if left == "" {
		return parsedHeader{}, false, nil
	}
	bracketStart := indexOutsideQuotes(left, '[')
	if bracketStart == -1 {
		return parsedHeader{}, false, nil
	}
	rest := left[bracketStart+1:]
	bracketOffset := indexOutsideQuotes(rest, ']')
	if bracketOffset == -1 {
		if strict {
			return parsedHeader{}, false, errors.New("missing closing bracket in array header")
		}
		return parsedHeader{}, false, nil
	}
	keyPart := left[:bracketStart]
	bracketSegment := rest[:bracketOffset]
	rawAfterBracket := rest[bracketOffset+1:]
	fieldSegment := strings.TrimSpace(rawAfterBracket)

	if strict && rawAfterBracket != "" {
		if strings.TrimSpace(rawAfterBracket) == "" {
			return parsedHeader{}, false, errors.New("whitespace between bracket segment and colon")
		}
		if rawAfterBracket[0] == ' ' && strings.HasPrefix(strings.TrimSpace(rawAfterBracket), "{") {
			return parsedHeader{}, false, errors.New("whitespace between bracket segment and fields segment")
		}
	}

	if fieldSegment != "" && (!strings.HasPrefix(fieldSegment, "{") || !strings.HasSuffix(fieldSegment, "}")) {
		if strict {
			return parsedHeader{}, false, errors.New("invalid field segment in array header")
		}
		return parsedHeader{}, false, nil
	}
	if strings.TrimSpace(fieldSegment) != "" && fieldSegment != "" {
		inner := strings.TrimSpace(fieldSegment[1 : len(fieldSegment)-1])
		if inner != "" && !strings.HasPrefix(fieldSegment, "{") {
			if strict {
				return parsedHeader{}, false, errors.New("invalid field segment in array header")
			}
			return parsedHeader{}, false, nil
		}
	}
	// Content between ] and {/: invalid in strict
	trimmedBetween := strings.TrimSpace(rest[bracketOffset+1:])
	if trimmedBetween != "" && !strings.HasPrefix(trimmedBetween, "{") {
		if strict {
			return parsedHeader{}, false, errors.New("invalid content between bracket and colon")
		}
		return parsedHeader{}, false, nil
	}

	header := parsedHeader{delimiter: DelimiterComma}
	header.hasKey = bracketStart > 0
	if strings.TrimSpace(keyPart) != "" {
		header.hasKey = true
		trimmedKey := strings.TrimSpace(keyPart)
		header.keyQuoted = strings.HasPrefix(trimmedKey, `"`)
		key, err := decodeKeyToken(trimmedKey, strict)
		if err != nil {
			if strict {
				return parsedHeader{}, false, err
			}
			return parsedHeader{}, false, nil
		}
		header.key = key
	}

	length, delim, err := parseBracketSegment(bracketSegment, strict)
	if err != nil {
		if strict {
			return parsedHeader{}, false, err
		}
		return parsedHeader{}, false, nil
	}
	header.length = length
	header.delimiter = delim

	if fieldSegment != "" {
		inner := fieldSegment[1 : len(fieldSegment)-1]
		if inner != "" {
			rawFields, err := SplitInlineValues(inner, delim.rune())
			if err != nil {
				return parsedHeader{}, false, err
			}
			fields := make([]string, 0, len(rawFields))
			for _, token := range rawFields {
				field, err := decodeKeyToken(token, strict)
				if err != nil {
					return parsedHeader{}, false, err
				}
				fields = append(fields, field)
			}
			header.fields = fields
		}
	}

	header.inlineValues = right
	return header, true, nil
}

func parseBracketSegment(segment string, strict bool) (int, Delimiter, error) {
	if strings.HasPrefix(segment, "#") {
		segment = segment[1:]
	}
	if segment == "" {
		return 0, DelimiterComma, errors.New("missing array length")
	}
	var digits strings.Builder
	delim := DelimiterComma
	for _, r := range segment {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
			continue
		}
		switch r {
		case '\t':
			delim = DelimiterTab
		case '|':
			delim = DelimiterPipe
		default:
			return 0, DelimiterComma, fmt.Errorf("invalid delimiter symbol %q", r)
		}
	}
	lengthStr := digits.String()
	if lengthStr == "" {
		return 0, DelimiterComma, errors.New("missing digits in array length")
	}
	if strict && len(lengthStr) > 1 && lengthStr[0] == '0' {
		return 0, DelimiterComma, errors.New("invalid array length with leading zeros")
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, DelimiterComma, err
	}
	return length, delim, nil
}

func (p *parser) splitKeyValue(content string) (string, string, bool, error) {
	colon := indexOutsideQuotes(content, ':')
	if colon == -1 {
		if !p.cfg.Strict {
			return strings.TrimSpace(content), "", false, nil
		}
		return "", "", false, errors.New("missing colon after key")
	}
	keyToken := strings.TrimSpace(content[:colon])
	valueToken := strings.TrimSpace(content[colon+1:])
	keyQuoted := strings.HasPrefix(keyToken, `"`)
	key, err := decodeKeyToken(keyToken, p.cfg.Strict)
	if err != nil {
		if !p.cfg.Strict {
			return keyToken, valueToken, keyQuoted, nil
		}
		return "", "", false, err
	}
	return key, valueToken, keyQuoted, nil
}

func decodeKeyToken(token string, strict bool) (string, error) {
	if token == "" {
		return "", errors.New("empty key")
	}
	if token[0] == '"' {
		return UnquoteString(token)
	}
	if !isValidUnquotedKey(token) {
		if strict {
			return "", fmt.Errorf("invalid unquoted key %q", token)
		}
		return token, nil
	}
	return token, nil
}

func decodePrimitiveToken(token string) (any, error) {
	if token == "" {
		return "", nil
	}
	if token[0] == '"' {
		return UnquoteString(token)
	}
	switch token {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null":
		return nil, nil
	}
	if hasForbiddenLeadingZeros(token) {
		return token, nil
	}
	if looksNumeric(token) {
		num, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return nil, err
		}
		if num == 0 {
			num = 0
		}
		return num, nil
	}
	return token, nil
}

func hasForbiddenLeadingZeros(token string) bool {
	if len(token) < 2 {
		return false
	}
	if token[0] == '-' {
		if len(token) > 2 && token[1] == '0' && token[2] >= '0' && token[2] <= '9' {
			if !strings.Contains(token, ".") && !strings.ContainsAny(token, "eE") {
				return true
			}
		}
		return false
	}
	if token[0] == '0' && token[1] >= '0' && token[1] <= '9' {
		if strings.Contains(token, ".") || strings.ContainsAny(token, "eE") {
			return false
		}
		return true
	}
	return false
}

func isKeyValue(content string) bool {
	return indexOutsideQuotes(content, ':') > 0
}

func indexOutsideQuotes(s string, target rune) int {
	inQuotes := false
	escaped := false
	for idx, r := range s {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inQuotes:
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
		case !inQuotes && r == target:
			return idx
		}
	}
	return -1
}

func (d Delimiter) rune() rune { return rune(d) }
