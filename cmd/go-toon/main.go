package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/abhishektakale/go-toon/toon"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "encode", "e":
		os.Exit(runEncode(os.Args[2:]))
	case "decode", "d":
		os.Exit(runDecode(os.Args[2:]))
	case "help", "-h", "--help":
		printUsage(os.Stdout)
	default:
		// Backward compatible: no subcommand means encode.
		os.Exit(runEncode(os.Args[1:]))
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  go-toon encode [flags]   # JSON stdin → TOON stdout")
	fmt.Fprintln(w, "  go-toon decode [flags]   # TOON stdin → JSON stdout")
	fmt.Fprintln(w, "  go-toon [flags]          # same as encode")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "encode flags:")
	fmt.Fprintln(w, "  -indent int           spaces per level (default 2)")
	fmt.Fprintln(w, "  -delimiter string     comma, tab, or pipe (default comma)")
	fmt.Fprintln(w, "  -length-markers       emit # length markers")
	fmt.Fprintln(w, "  -key-folding string   off or safe (default off)")
	fmt.Fprintln(w, "  -fast                 use fast JSON parse (unordered keys)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "decode flags:")
	fmt.Fprintln(w, "  -indent int           expected spaces per level (default 2)")
	fmt.Fprintln(w, "  -delimiter string     comma, tab, or pipe (default comma)")
	fmt.Fprintln(w, "  -strict               enable strict validation (default true)")
	fmt.Fprintln(w, "  -no-strict            disable strict validation")
	fmt.Fprintln(w, "  -expand-paths string  off or safe (default off)")
	fmt.Fprintln(w, "  -pretty               pretty-print JSON output")
}

func runEncode(args []string) int {
	fs := newEncodeFlags()
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		return 2
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		return 1
	}

	delim, err := parseDelimiterFlag(fs.delimiter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		return 2
	}

	opts := toon.EncodeOptions{
		Indent:        fs.indent,
		Delimiter:     delim,
		LengthMarkers: fs.lengthMarkers,
	}
	if fs.keyFolding != "" {
		opts.KeyFolding = toon.KeyFolding(fs.keyFolding)
	}

	var out []byte
	if fs.fast {
		out, err = toon.EncodeJSONFast(data, opts)
	} else {
		out, err = toon.EncodeJSON(data, opts)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		return 1
	}

	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func runDecode(args []string) int {
	fs := newDecodeFlags()
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		return 2
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		return 1
	}

	delim, err := parseDelimiterFlag(fs.delimiter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		return 2
	}

	opts := toon.DecodeOptions{
		Indent:      fs.indent,
		Strict:      fs.strict,
		Delimiter:   delim,
		ExpandPaths: fs.expandPaths,
	}

	var out []byte
	if fs.pretty {
		v, err := toon.Decode(data, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode: %v\n", err)
			return 1
		}
		out, err = json.MarshalIndent(v, "", strings.Repeat(" ", fs.indent))
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal json: %v\n", err)
			return 1
		}
	} else {
		out, err = toon.DecodeToJSON(data, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode: %v\n", err)
			return 1
		}
	}

	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func parseDelimiterFlag(s string) (toon.Delimiter, error) {
	switch s {
	case "comma", ",":
		return toon.DelimiterComma, nil
	case "tab", "\\t":
		return toon.DelimiterTab, nil
	case "pipe", "|":
		return toon.DelimiterPipe, nil
	default:
		return 0, fmt.Errorf("unknown delimiter %q (use comma, tab, or pipe)", s)
	}
}
