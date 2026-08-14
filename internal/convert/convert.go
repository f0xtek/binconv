// Package convert parses and formats arbitrary-precision integers across
// binary, octal, decimal and hexadecimal bases.
package convert

import (
	"fmt"
	"math/big"
	"strings"
)

// Base describes one of the supported numeric bases.
type Base struct {
	Name   string // canonical name: "binary", "octal", "decimal", "hexadecimal"
	Radix  int    // 2, 8, 10, 16
	Prefix string // "0b", "0o", "", "0x"
}

var (
	Binary      = Base{Name: "binary", Radix: 2, Prefix: "0b"}
	Octal       = Base{Name: "octal", Radix: 8, Prefix: "0o"}
	Decimal     = Base{Name: "decimal", Radix: 10, Prefix: ""}
	Hexadecimal = Base{Name: "hexadecimal", Radix: 16, Prefix: "0x"}
)

// aliases maps every accepted spelling (lowercase) to its canonical Base.
var aliases = map[string]Base{
	"binary": Binary, "bin": Binary, "b": Binary, "2": Binary,
	"octal": Octal, "oct": Octal, "o": Octal, "8": Octal,
	"decimal": Decimal, "dec": Decimal, "d": Decimal, "10": Decimal,
	"hexadecimal": Hexadecimal, "hex": Hexadecimal, "x": Hexadecimal, "h": Hexadecimal, "16": Hexadecimal,
}

// ParseBase resolves a base name or alias (case-insensitive) to a Base.
func ParseBase(s string) (Base, error) {
	if b, ok := aliases[strings.ToLower(s)]; ok {
		return b, nil
	}
	return Base{}, fmt.Errorf("unknown base %q (valid: binary, octal, decimal, hexadecimal)", s)
}

// Options controls how Format renders its output.
type Options struct {
	Prefix bool // prepend the base's canonical prefix (0b/0o/0x); no-op for decimal
	Upper  bool // uppercase hexadecimal digits
}

// Parse interprets s as a number in the given base and returns its value.
//
// A leading '+' or '-' sign is honoured. If the remaining digits carry a
// prefix matching from's own prefix (e.g. "0x" for hexadecimal), that
// prefix is stripped. A prefix belonging to a different base is left as
// part of the digit string — this is deliberate, since e.g. "0b1" is a
// legitimate hexadecimal value (177), not a binary literal in a
// hexadecimal context.
func Parse(s string, from Base) (*big.Int, error) {
	orig := s
	sign := ""
	trimmed := s
	if len(trimmed) > 0 && (trimmed[0] == '+' || trimmed[0] == '-') {
		if trimmed[0] == '-' {
			sign = "-"
		}
		trimmed = trimmed[1:]
	}

	digits := trimmed
	foreignPrefix := ""
	if from.Prefix != "" && hasPrefixFold(trimmed, from.Prefix) {
		digits = trimmed[len(from.Prefix):]
	} else if p := detectForeignPrefix(trimmed, from); p != "" {
		foreignPrefix = p
	}

	n, ok := new(big.Int).SetString(sign+digits, from.Radix)
	if !ok {
		if foreignPrefix != "" {
			baseName := prefixBaseName(foreignPrefix)
			return nil, fmt.Errorf("invalid %s value %q (%s is a %s prefix — did you mean to convert from %s?)",
				from.Name, orig, foreignPrefix, baseName, baseName)
		}
		return nil, fmt.Errorf("invalid %s value %q", from.Name, orig)
	}
	return n, nil
}

// hasPrefixFold reports whether s starts with prefix, ignoring case.
func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

// detectForeignPrefix returns the recognized base prefix at the start of s,
// if any, provided it does not belong to from.
func detectForeignPrefix(s string, from Base) string {
	for _, b := range []Base{Binary, Octal, Hexadecimal} {
		if b.Prefix == "" || b.Radix == from.Radix {
			continue
		}
		if hasPrefixFold(s, b.Prefix) {
			return s[:len(b.Prefix)]
		}
	}
	return ""
}

func prefixBaseName(prefix string) string {
	switch strings.ToLower(prefix) {
	case "0b":
		return "binary"
	case "0o":
		return "octal"
	case "0x":
		return "hexadecimal"
	}
	return ""
}

// Format renders n in the given base, applying the requested Options.
func Format(n *big.Int, to Base, opts Options) string {
	text := n.Text(to.Radix)
	if opts.Upper {
		text = strings.ToUpper(text)
	}
	if opts.Prefix && to.Prefix != "" {
		if strings.HasPrefix(text, "-") {
			text = "-" + to.Prefix + text[1:]
		} else {
			text = to.Prefix + text
		}
	}
	return text
}
