package toon

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

func formatNumber(v any) string {
	switch n := v.(type) {
	case json.Number:
		return formatJSONNumber(n)
	case float64:
		return formatFloat64(n)
	case float32:
		return formatFloat64(float64(n))
	case int:
		return strconv.Itoa(n)
	case int64:
		return strconv.FormatInt(n, 10)
	case int32:
		return strconv.FormatInt(int64(n), 10)
	default:
		return "null"
	}
}

func formatJSONNumber(n json.Number) string {
	if i, err := n.Int64(); err == nil {
		if f, err := n.Float64(); err == nil && float64(i) == f {
			return strconv.FormatInt(i, 10)
		}
	}
	f, err := n.Float64()
	if err != nil {
		return string(n)
	}
	return formatFloat64(f)
}

func formatFloat64(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "null"
	}
	if f == 0 {
		return "0"
	}

	abs := math.Abs(f)
	if abs >= 1e-6 && abs < 1e21 {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
		if s == "-0" {
			return "0"
		}
		return s
	}

	s := strconv.FormatFloat(f, 'e', -1, 64)
	if idx := strings.Index(s, "e"); idx >= 0 {
		exp := s[idx+1:]
		if len(exp) > 0 && exp[0] != '+' && exp[0] != '-' {
			s = s[:idx+1] + "+" + exp
		}
	}
	return s
}
