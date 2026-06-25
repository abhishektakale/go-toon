package toon

import (
	"io"
	"unicode/utf8"
)

func needsQuoting(s string, delim Delimiter) bool {
	if s == "" {
		return true
	}
	if len(s) > 0 && (s[0] <= ' ' || s[len(s)-1] <= ' ') {
		return true
	}
	switch s {
	case "true", "false", "null":
		return true
	}
	if looksNumeric(s) {
		return true
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ':', '"', '[', '{', '\\':
			return true
		}
	}
	if len(s) > 0 && s[0] == '-' {
		return true
	}
	for i := 0; i < len(s); i++ {
		if s[i] <= 0x1F {
			return true
		}
	}
	if delim != 0 && byte(delim) < utf8.RuneSelf {
		for i := 0; i < len(s); i++ {
			if s[i] == byte(delim) {
				return true
			}
		}
	}
	return false
}

func writeEscapedString(w io.Writer, s string) {
	for _, r := range s {
		switch r {
		case '\\':
			io.WriteString(w, `\\`)
		case '"':
			io.WriteString(w, `\"`)
		case '\n':
			io.WriteString(w, `\n`)
		case '\r':
			io.WriteString(w, `\r`)
		case '\t':
			io.WriteString(w, `\t`)
		default:
			if r <= 0x1F {
				io.WriteString(w, `\u`)
				io.WriteString(w, hex4(r))
			} else {
				io.WriteString(w, string(r))
			}
		}
	}
}

func hex4(r rune) string {
	const digits = "0123456789abcdef"
	return string([]byte{
		digits[(r>>12)&0xF],
		digits[(r>>8)&0xF],
		digits[(r>>4)&0xF],
		digits[r&0xF],
	})
}

func writeQuotedString(w io.Writer, s string) {
	w.Write([]byte{'"'})
	writeEscapedString(w, s)
	w.Write([]byte{'"'})
}

func writeKey(w io.Writer, key string) {
	if !isValidUnquotedKey(key) {
		writeQuotedString(w, key)
		return
	}
	io.WriteString(w, key)
}

func writeQuotedPrimitive(w io.Writer, s string, delim Delimiter) {
	if needsQuoting(s, delim) {
		writeQuotedString(w, s)
		return
	}
	io.WriteString(w, s)
}
