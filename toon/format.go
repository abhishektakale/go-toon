package toon

import "unicode"

func looksNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	i := 0
	if s[0] == '-' {
		i++
		if i == len(s) {
			return false
		}
	}
	digits := 0
	for i < len(s) && isDigit(s[i]) {
		i++
		digits++
	}
	if digits == 0 {
		return false
	}
	if i < len(s) && s[i] == '.' {
		i++
		if i == len(s) || !isDigit(s[i]) {
			return false
		}
		for i < len(s) && isDigit(s[i]) {
			i++
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if i == len(s) || !isDigit(s[i]) {
			return false
		}
		for i < len(s) && isDigit(s[i]) {
			i++
		}
	}
	return i == len(s)
}

func isValidUnquotedKey(key string) bool {
	if key == "" {
		return false
	}
	for pos, r := range key {
		if pos == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
