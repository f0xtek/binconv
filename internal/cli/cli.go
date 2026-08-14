// Package cli implements the binconv command-line interface.
package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/f0xtek/binconv/internal/convert"
)

const usageText = `binconv — convert numbers between binary, octal, decimal and hexadecimal

Usage:
  binconv [flags] <from> <to> [value ...]
  ... | binconv [flags] <from> <to>

Bases (case-insensitive):
  binary        bin, b, 2
  octal         oct, o, 8
  decimal       dec, d, 10
  hexadecimal   hex, x, h, 16

Flags:
      --no-prefix  omit the 0b/0o/0x prefix (on by default; no effect on decimal)
  -u, --upper      uppercase hexadecimal digits
  -h, --help       show this help
  -V, --version    show version

Examples:
  binconv decimal hexadecimal 1337     0x539
  binconv dec hex -u 1337              0x539
  binconv dec hex --no-prefix 1337     539
  binconv hex dec 0x539 ff 10          one result per line
  printf '1\n10\n11\n' | binconv bin dec
`

// version is overridable at build time via -ldflags "-X .../cli.version=...".
var version = "dev"

// knownFlags maps every recognized flag spelling to true, used to decide
// whether a "-x" argument is a flag or a (possibly negative) value.
var knownFlags = map[string]bool{
	"--no-prefix": true,
	"-u":          true, "--upper": true,
	"-h": true, "--help": true,
	"-V": true, "--version": true,
}

// Run executes the binconv CLI against the given argv (excluding the
// program name) and I/O streams, returning the process exit code.
func Run(argv []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flagArgs, positional := permute(argv)

	fs := flag.NewFlagSet("binconv", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usageText) }

	var noPrefix, upper, help, showVersion bool
	fs.BoolVar(&noPrefix, "no-prefix", false, "")
	fs.BoolVar(&upper, "upper", false, "")
	fs.BoolVar(&upper, "u", false, "")
	fs.BoolVar(&help, "help", false, "")
	fs.BoolVar(&help, "h", false, "")
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&showVersion, "V", false, "")

	if err := fs.Parse(flagArgs); err != nil {
		return 2
	}

	if help {
		fmt.Fprint(stdout, usageText)
		return 0
	}
	if showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	if len(positional) < 2 {
		fmt.Fprintln(stderr, "binconv: missing <from> and/or <to> base")
		fmt.Fprint(stderr, usageText)
		return 2
	}

	fromArg, toArg, values := positional[0], positional[1], positional[2:]

	from, err := convert.ParseBase(fromArg)
	if err != nil {
		fmt.Fprintf(stderr, "binconv: %s\n", err)
		return 2
	}
	to, err := convert.ParseBase(toArg)
	if err != nil {
		fmt.Fprintf(stderr, "binconv: %s\n", err)
		return 2
	}

	opts := convert.Options{Prefix: !noPrefix, Upper: upper}

	if len(values) == 1 && values[0] == "-" {
		values = nil
	}

	if len(values) == 0 {
		if isTerminal(stdin) {
			fmt.Fprintln(stderr, "binconv: no values given and stdin is a terminal")
			fmt.Fprint(stderr, usageText)
			return 2
		}
		return convertStream(stdin, stdout, stderr, from, to, opts)
	}

	return convertValues(values, stdout, stderr, from, to, opts)
}

// permute splits argv into recognized flag arguments and positional
// arguments, so that flags may appear anywhere on the command line
// (stdlib flag.FlagSet otherwise stops parsing at the first positional).
// A "--" terminator forces everything after it to be treated as
// positional. An argument is only ever classified as a flag if it matches
// a known flag spelling — this keeps negative values like "-2a" intact.
func permute(argv []string) (flags, positional []string) {
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if a == "--" {
			positional = append(positional, argv[i+1:]...)
			break
		}
		if knownFlags[a] {
			flags = append(flags, a)
			continue
		}
		positional = append(positional, a)
	}
	return flags, positional
}

// isTerminal reports whether r is an interactive terminal (best effort;
// only meaningful when r is *os.File, e.g. os.Stdin).
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// convertValues converts each value in turn, writing successes to stdout
// and failures to stderr, returning 1 if any conversion failed.
func convertValues(values []string, stdout, stderr io.Writer, from, to convert.Base, opts convert.Options) int {
	failed := false
	for _, v := range values {
		n, err := convert.Parse(v, from)
		if err != nil {
			fmt.Fprintf(stderr, "binconv: %s\n", err)
			failed = true
			continue
		}
		fmt.Fprintln(stdout, convert.Format(n, to, opts))
	}
	if failed {
		return 1
	}
	return 0
}

// convertStream reads one value per line from r, converting each and
// skipping blank lines, returning 1 if any conversion failed.
func convertStream(r io.Reader, stdout, stderr io.Writer, from, to convert.Base, opts convert.Options) int {
	failed := false
	scanner := bufio.NewScanner(r)
	line := 0
	for scanner.Scan() {
		line++
		v := strings.TrimSpace(scanner.Text())
		if v == "" {
			continue
		}
		n, err := convert.Parse(v, from)
		if err != nil {
			fmt.Fprintf(stderr, "binconv: line %d: %s\n", line, err)
			failed = true
			continue
		}
		fmt.Fprintln(stdout, convert.Format(n, to, opts))
	}
	if failed {
		return 1
	}
	return 0
}
