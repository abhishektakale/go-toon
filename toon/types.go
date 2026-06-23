package toon

// Delimiter separates values in inline arrays and tabular rows.
type Delimiter rune

const (
	DelimiterComma Delimiter = ','
	DelimiterTab   Delimiter = '\t'
	DelimiterPipe  Delimiter = '|'
)

// Field is one key/value pair. Order matters.
type Field struct {
	Key   string
	Value any
}

// Object is an ordered list of fields (not a map).
type Object []Field

// EncodeOptions tweaks output formatting.
type EncodeOptions struct {
	Indent        int
	Delimiter     Delimiter
	LengthMarkers bool
}

func (o EncodeOptions) withDefaults() EncodeOptions {
	if o.Indent <= 0 {
		o.Indent = 2
	}
	if o.Delimiter == 0 {
		o.Delimiter = DelimiterComma
	}
	return o
}
