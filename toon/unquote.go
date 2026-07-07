package toon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// UnquoteString removes surrounding quotes and unescapes TOON strings per §7.1.
func UnquoteString(token string) (string, error) {
	if len(token) < 2 || token[0] != '"' || token[len(token)-1] != '"' {
		return "", errors.New("invalid quoted string")
	}
	var b strings.Builder
	b.Grow(len(token) - 2)
	escaped := false
	for i := 1; i < len(token)-1; i++ {
		ch := token[i]
		if escaped {
			switch ch {
			case '\\', '"':
				b.WriteByte(ch)
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				if i+4 >= len(token)-1 {
					return "", errors.New("invalid unicode escape")
				}
				hex := token[i+1 : i+5]
				if !isHex4(hex) {
					return "", errors.New("invalid unicode escape")
				}
				cp, err := strconv.ParseUint(hex, 16, 16)
				if err != nil {
					return "", err
				}
				if cp >= 0xD800 && cp <= 0xDFFF {
					return "", errors.New("lone surrogate in unicode escape")
				}
				b.WriteRune(rune(cp))
				i += 4
			default:
				return "", fmt.Errorf("invalid escape sequence \\%c", ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		b.WriteByte(ch)
	}
	if escaped {
		return "", errors.New("unterminated escape sequence")
	}
	if !utf8.ValidString(b.String()) {
		return "", errors.New("invalid UTF-8 in string")
	}
	return b.String(), nil
}

func isHex4(s string) bool {
	if len(s) != 4 {
		return false
	}
	for i := 0; i < 4; i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// SplitInlineValues tokenizes a delimiter-separated list, respecting quoted segments.
// Each token is sliced directly out of segment rather than rebuilt rune by rune;
// quotes and escapes are preserved in place and unescaped later by
// decodePrimitiveToken/UnquoteString.
func SplitInlineValues(segment string, delimiter rune) ([]string, error) {
	if strings.TrimSpace(segment) == "" {
		return nil, nil
	}
	var tokens []string
	inQuotes := false
	escaped := false
	start := 0

	for i, r := range segment {
		if escaped {
			escaped = false
			continue
		}
		switch {
		case r == '\\' && inQuotes:
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
		case r == delimiter && !inQuotes:
			tokens = append(tokens, strings.TrimSpace(segment[start:i]))
			start = i + utf8.RuneLen(r)
		}
	}
	if inQuotes {
		return nil, errors.New("unterminated string in delimited values")
	}
	tokens = append(tokens, strings.TrimSpace(segment[start:]))
	return tokens, nil
}
