package convert

import (
	"math/big"
	"testing"
)

func TestParseBase(t *testing.T) {
	cases := map[string]Base{
		"binary": Binary, "BIN": Binary, "b": Binary, "2": Binary,
		"octal": Octal, "OCT": Octal, "o": Octal, "8": Octal,
		"decimal": Decimal, "Dec": Decimal, "d": Decimal, "10": Decimal,
		"hexadecimal": Hexadecimal, "HEX": Hexadecimal, "x": Hexadecimal, "h": Hexadecimal, "16": Hexadecimal,
	}
	for in, want := range cases {
		got, err := ParseBase(in)
		if err != nil {
			t.Errorf("ParseBase(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseBase(%q) = %+v, want %+v", in, got, want)
		}
	}

	if _, err := ParseBase("nope"); err == nil {
		t.Error("ParseBase(\"nope\") expected error, got nil")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		from    Base
		want    string // decimal string of expected value
		wantErr bool
	}{
		{"plain decimal", "1337", Decimal, "1337", false},
		{"plain binary", "1010", Binary, "10", false},
		{"plain octal", "17", Octal, "15", false},
		{"plain hex", "539", Hexadecimal, "1337", false},
		{"matching prefix stripped", "0x539", Hexadecimal, "1337", false},
		{"matching prefix stripped case-insensitive", "0X539", Hexadecimal, "1337", false},
		{"matching binary prefix", "0b1010", Binary, "10", false},
		{"matching octal prefix", "0o17", Octal, "15", false},
		{"foreign prefix treated as hex digits", "0b1", Hexadecimal, "177", false},
		{"negative decimal", "-42", Decimal, "-42", false},
		{"negative hex", "-2a", Hexadecimal, "-42", false},
		{"negative with matching prefix", "-0x2a", Hexadecimal, "-42", false},
		{"invalid digit for base", "129", Binary, "", true},
		{"empty string", "", Decimal, "", true},
		{"foreign prefix invalid in decimal", "0x539", Decimal, "", true},
		{"large value beyond uint64", "123456789012345678901234567890", Decimal, "123456789012345678901234567890", false},
		{"zero", "0", Decimal, "0", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.in, tc.from)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q, %v) expected error, got nil (value=%v)", tc.in, tc.from, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q, %v) unexpected error: %v", tc.in, tc.from, err)
			}
			want, ok := new(big.Int).SetString(tc.want, 10)
			if !ok {
				t.Fatalf("test bug: bad want value %q", tc.want)
			}
			if got.Cmp(want) != 0 {
				t.Errorf("Parse(%q, %v) = %v, want %v", tc.in, tc.from, got, want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	n := big.NewInt(1337)
	neg := big.NewInt(-42)
	zero := big.NewInt(0)

	tests := []struct {
		name string
		n    *big.Int
		to   Base
		opts Options
		want string
	}{
		{"decimal bare", n, Decimal, Options{}, "1337"},
		{"hex bare lowercase", n, Hexadecimal, Options{}, "539"},
		{"hex upper", n, Hexadecimal, Options{Upper: true}, "539"},
		{"hex upper with letters", big.NewInt(0xabc), Hexadecimal, Options{Upper: true}, "ABC"},
		{"hex with prefix", n, Hexadecimal, Options{Prefix: true}, "0x539"},
		{"hex with prefix and upper", big.NewInt(0xabc), Hexadecimal, Options{Prefix: true, Upper: true}, "0xABC"},
		{"decimal prefix is no-op", n, Decimal, Options{Prefix: true}, "1337"},
		{"binary with prefix", big.NewInt(10), Binary, Options{Prefix: true}, "0b1010"},
		{"octal with prefix", big.NewInt(15), Octal, Options{Prefix: true}, "0o17"},
		{"negative with prefix placed after sign", neg, Hexadecimal, Options{Prefix: true}, "-0x2a"},
		{"negative without prefix", neg, Hexadecimal, Options{}, "-2a"},
		{"zero", zero, Hexadecimal, Options{Prefix: true}, "0x0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Format(tc.n, tc.to, tc.opts)
			if got != tc.want {
				t.Errorf("Format(%v, %v, %+v) = %q, want %q", tc.n, tc.to, tc.opts, got, tc.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	bases := []Base{Binary, Octal, Decimal, Hexadecimal}
	values := []int64{0, 1, 42, 1337, -42, -1337, 1000000007}

	for _, v := range values {
		start := big.NewInt(v)
		for _, from := range bases {
			s := Format(start, from, Options{})
			got, err := Parse(s, from)
			if err != nil {
				t.Fatalf("round trip: Parse(Format(%v, %v)) failed: %v", v, from, err)
			}
			if got.Cmp(start) != 0 {
				t.Errorf("round trip: value=%v base=%v: got %v, want %v", v, from.Name, got, start)
			}
		}
	}
}
