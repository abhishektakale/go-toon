package toon

import "fmt"

// ParseError describes a TOON syntax or validation error at a line number.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("toon: line %d: %s", e.Line, e.Msg)
	}
	return "toon: " + e.Msg
}

func errorAt(line int, msg string) error {
	return &ParseError{Line: line, Msg: msg}
}

func errorAtf(line int, format string, args ...any) error {
	return &ParseError{Line: line, Msg: fmt.Sprintf(format, args...)}
}

func errorWrap(line int, err error) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(*ParseError); ok {
		if pe.Line > 0 {
			return pe
		}
		return &ParseError{Line: line, Msg: pe.Msg}
	}
	return &ParseError{Line: line, Msg: err.Error()}
}
