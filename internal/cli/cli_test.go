package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(t *testing.T, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = Run(args, strings.NewReader(stdin), &out, &errBuf)
	return out.String(), errBuf.String(), code
}

func TestWorkedExample(t *testing.T) {
	// Prefixes are on by default, so hexadecimal output carries "0x".
	out, errOut, code := run(t, "", "decimal", "hexadecimal", "1337")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, errOut)
	}
	if out != "0x539\n" {
		t.Errorf("stdout = %q, want %q", out, "0x539\n")
	}
}

func TestPrefixDefaultOnAcrossBases(t *testing.T) {
	tests := []struct {
		to   string
		val  string
		want string
	}{
		{"bin", "10", "0b1010\n"},
		{"oct", "15", "0o17\n"},
		{"hex", "1337", "0x539\n"},
		{"dec", "1337", "1337\n"}, // decimal has no prefix to add
	}
	for _, tc := range tests {
		out, errOut, code := run(t, "", "dec", tc.to, tc.val)
		if code != 0 {
			t.Fatalf("to=%s: exit code = %d, want 0 (stderr=%q)", tc.to, code, errOut)
		}
		if out != tc.want {
			t.Errorf("to=%s: stdout = %q, want %q", tc.to, out, tc.want)
		}
	}
}

func TestNoPrefixFlag(t *testing.T) {
	out, _, code := run(t, "", "dec", "hex", "--no-prefix", "1337")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if out != "539\n" {
		t.Errorf("stdout = %q, want %q", out, "539\n")
	}
}

func TestFlagsPermuteAfterPositionals(t *testing.T) {
	// --no-prefix and -u should both apply even though they appear after
	// the positional arguments, and uppercase should affect hex letters.
	out, _, code := run(t, "", "dec", "hex", "--no-prefix", "-u", "2748")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if out != "ABC\n" {
		t.Errorf("stdout = %q, want %q", out, "ABC\n")
	}
}

func TestNegativeHexValueNotTreatedAsFlag(t *testing.T) {
	out, errOut, code := run(t, "", "hex", "dec", "-2a")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, errOut)
	}
	if out != "-42\n" {
		t.Errorf("stdout = %q, want %q", out, "-42\n")
	}
}

func TestMultipleValues(t *testing.T) {
	out, _, code := run(t, "", "hex", "dec", "0x539", "ff", "10")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "1337\n255\n16\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestStdinMultiLine(t *testing.T) {
	out, _, code := run(t, "1\n10\n11\n", "bin", "dec")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "1\n2\n3\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestStdinBlankLinesSkipped(t *testing.T) {
	out, _, code := run(t, "1\n\n10\n", "bin", "dec")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "1\n2\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
}

func TestMixedGoodBadBatchArgs(t *testing.T) {
	out, errOut, code := run(t, "", "bin", "dec", "1", "129", "11")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	want := "1\n3\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	if !strings.Contains(errOut, "129") {
		t.Errorf("stderr = %q, want mention of failing value 129", errOut)
	}
}

func TestMixedGoodBadBatchStdin(t *testing.T) {
	out, errOut, code := run(t, "1\n129\n11\n", "bin", "dec")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	want := "1\n3\n"
	if out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	if !strings.Contains(errOut, "line 2") {
		t.Errorf("stderr = %q, want mention of line 2", errOut)
	}
}

func TestBadBaseName(t *testing.T) {
	_, errOut, code := run(t, "", "nope", "hex", "1")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errOut, "nope") {
		t.Errorf("stderr = %q, want mention of bad base %q", errOut, "nope")
	}
}

func TestNoValuesEmptyNonTerminalStdin(t *testing.T) {
	// In tests, stdin is a strings.Reader (not *os.File), so it is never
	// treated as a terminal. No values on the command line plus empty
	// piped stdin means: read zero lines, convert nothing, succeed.
	out, errOut, code := run(t, "", "dec", "hex")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", code, errOut)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
}

func TestMissingFromAndTo(t *testing.T) {
	_, errOut, code := run(t, "", "dec")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errOut, "missing") {
		t.Errorf("stderr = %q, want mention of missing bases", errOut)
	}
}

func TestForeignPrefixHintInDecimal(t *testing.T) {
	_, errOut, code := run(t, "", "dec", "hex", "0x539")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "hexadecimal") {
		t.Errorf("stderr = %q, want hint mentioning hexadecimal prefix", errOut)
	}
}

func TestHelpFlag(t *testing.T) {
	out, _, code := run(t, "", "--help")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("stdout = %q, want usage text", out)
	}
}

func TestVersionFlag(t *testing.T) {
	out, _, code := run(t, "", "-V")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected version output")
	}
}

func TestDashMeansStdin(t *testing.T) {
	out, _, code := run(t, "1010\n", "bin", "dec", "-")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if out != "10\n" {
		t.Errorf("stdout = %q, want %q", out, "10\n")
	}
}

func TestBigIntValue(t *testing.T) {
	out, _, code := run(t, "", "dec", "hex", "123456789012345678901234567890")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.HasPrefix(out, "0x") {
		t.Errorf("stdout = %q, want 0x-prefixed value by default", out)
	}
}
