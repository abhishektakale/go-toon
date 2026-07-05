package main

import "flag"

type encodeFlags struct {
	fs            *flag.FlagSet
	indent        int
	delimiter     string
	lengthMarkers bool
	keyFolding    string
	fast          bool
}

func newEncodeFlags() *encodeFlags {
	f := &encodeFlags{fs: flag.NewFlagSet("encode", flag.ContinueOnError)}
	f.fs.IntVar(&f.indent, "indent", 2, "spaces per indentation level")
	f.fs.StringVar(&f.delimiter, "delimiter", "comma", "array delimiter: comma, tab, or pipe")
	f.fs.BoolVar(&f.lengthMarkers, "length-markers", false, "prefix array lengths with #")
	f.fs.StringVar(&f.keyFolding, "key-folding", "", "key folding mode: off or safe")
	f.fs.BoolVar(&f.fast, "fast", false, "use fast JSON parse (unordered keys)")
	return f
}

func (f *encodeFlags) Parse(args []string) error {
	return f.fs.Parse(args)
}

type decodeFlags struct {
	fs          *flag.FlagSet
	indent      int
	delimiter   string
	strict      bool
	noStrict    bool
	expandPaths string
	pretty      bool
}

func newDecodeFlags() *decodeFlags {
	f := &decodeFlags{fs: flag.NewFlagSet("decode", flag.ContinueOnError)}
	f.fs.IntVar(&f.indent, "indent", 2, "expected spaces per indentation level")
	f.fs.StringVar(&f.delimiter, "delimiter", "comma", "array delimiter: comma, tab, or pipe")
	f.fs.BoolVar(&f.strict, "strict", true, "enable strict validation")
	f.fs.BoolVar(&f.noStrict, "no-strict", false, "disable strict validation")
	f.fs.StringVar(&f.expandPaths, "expand-paths", "off", "path expansion: off or safe")
	f.fs.BoolVar(&f.pretty, "pretty", false, "pretty-print JSON output")
	return f
}

func (f *decodeFlags) Parse(args []string) error {
	if err := f.fs.Parse(args); err != nil {
		return err
	}
	if f.noStrict {
		f.strict = false
	}
	return nil
}
