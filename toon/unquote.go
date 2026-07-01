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
func SplitInlineValues(segment string, delimiter rune) ([]string, error) {
	if strings.TrimSpace(segment) == "" {
		return nil, nil
	}
	var tokens []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range segment {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && inQuotes:
			current.WriteRune(r)
			escaped = true
		case r == '"':
			current.WriteRune(r)
			inQuotes = !inQuotes
		case r == delimiter && !inQuotes:
			tokens = append(tokens, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if inQuotes {
		return nil, errors.New("unterminated string in delimited values")
	}
	tokens = append(tokens, strings.TrimSpace(current.String()))
	return tokens, nil
}
