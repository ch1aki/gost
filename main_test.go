package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	st "github.com/ch1aki/gost/lib"
	"github.com/google/go-cmp/cmp"
)

func TestHelpExitsZero(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			help := captureOS(t)
			code := cli([]string{arg}, strings.NewReader(""), io.Discard, io.Discard)
			text := help()
			if code != 0 {
				t.Fatalf("cli(%s) = %d, want 0\n%s", arg, code, text)
			}
			if !strings.Contains(text, "fixed-point decimal") || strings.Contains(text, "outpuut") || strings.Contains(text, "default: %g") {
				t.Fatalf("help text = %q", text)
			}
		})
	}
}

func TestInvalidFormatExits(t *testing.T) {
	for _, format := range []string{"%d", "%s", "hello"} {
		t.Run(format, func(t *testing.T) {
			restore := captureOS(t)
			defer restore()
			var stdout, stderr bytes.Buffer
			code := cli([]string{"-f", format}, strings.NewReader("1\n2\n"), &stdout, &stderr)
			if code != 1 {
				t.Fatalf("cli() = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if diff := cmp.Diff(fmt.Sprintf("invalid format %q\n", format), stderr.String()); diff != "" {
				t.Errorf("stderr mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDefaultFormatFixedPoint(t *testing.T) {
	restore := captureOS(t)
	defer restore()
	var b strings.Builder
	for i := 1; i <= 10000; i++ {
		fmt.Fprintf(&b, "%d\n", i)
	}
	var stdout, stderr bytes.Buffer
	code := cli([]string{"--sum", "--max"}, strings.NewReader(b.String()), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("cli() = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "50005000") || !strings.Contains(out, "10000") {
		t.Fatalf("stdout = %q, want fixed-point sum and max", out)
	}
	if strings.Contains(out, "e+") || strings.Contains(out, "e-") {
		t.Fatalf("stdout = %q, default format used exponent", out)
	}

	stdout.Reset()
	stderr.Reset()
	code = cli([]string{"--max"}, strings.NewReader("0.0000001\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("cli() = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "0.0000001") || strings.Contains(stdout.String(), "e-") {
		t.Fatalf("stdout = %q, want fixed-point small value", stdout.String())
	}
}

func TestFormatFlag(t *testing.T) {
	restore := captureOS(t)
	defer restore()
	var stdout, stderr bytes.Buffer
	code := cli([]string{"-f", "%.2f", "-N", "-m"}, strings.NewReader("1\n2\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("cli() = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "1.50") {
		t.Fatalf("stdout = %q, want mean 1.50", out)
	}
	if strings.Contains(out, "2.00") {
		t.Fatalf("stdout = %q, N must stay an integer", out)
	}
}

// captureOS redirects process stdout and stderr, where go-flags writes help
// and parse errors, and returns a function that restores them and yields stdout.
func captureOS(t *testing.T) func() string {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	re, we, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = w, we
	return func() string {
		w.Close()
		we.Close()
		os.Stdout, os.Stderr = oldOut, oldErr
		out, _ := io.ReadAll(r)
		io.Copy(io.Discard, re)
		return string(out)
	}
}

func TestReadInput(t *testing.T) {
	long := strings.Repeat("1", bufio.MaxScanTokenSize) + "\n"

	tests := []struct {
		name    string
		in      string
		want    []float64
		wantErr string
	}{
		{
			name: "plain",
			in:   "1\n2\n3\n",
			want: []float64{1, 2, 3},
		},
		{
			name: "trim and skip blanks",
			in:   " 1 \n\n \t \n2\n",
			want: []float64{1, 2},
		},
		{
			name: "negatives",
			in:   " -2 \n-1\n0\n",
			want: []float64{-2, -1, 0},
		},
		{
			name: "crlf",
			in:   "1\r\n2\r\n",
			want: []float64{1, 2},
		},
		{
			name:    "invalid",
			in:      "1\n\n x \n3\n",
			want:    []float64{1},
			wantErr: "stdin:3: invalid number \"x\"",
		},
		{
			name:    "nan",
			in:      "NaN\n",
			wantErr: "stdin:1: unsupported value \"NaN\"",
		},
		{
			name:    "nan lowercase after value",
			in:      "1\nnan\n",
			want:    []float64{1},
			wantErr: "stdin:2: unsupported value \"nan\"",
		},
		{
			name:    "inf",
			in:      "+Inf\n",
			wantErr: "stdin:1: unsupported value \"+Inf\"",
		},
		{
			name:    "negative inf",
			in:      "-Inf\n",
			wantErr: "stdin:1: unsupported value \"-Inf\"",
		},
		{
			name:    "infinity word",
			in:      " Infinity \n",
			wantErr: "stdin:1: unsupported value \"Infinity\"",
		},
		{
			name:    "overflow",
			in:      "1e309\n",
			wantErr: "stdin:1: unsupported value \"1e309\"",
		},
		{
			name:    "line too long",
			in:      long,
			wantErr: "stdin:1: line too long",
		},
		{
			name:    "line too long after blank",
			in:      "1\n\n" + long,
			want:    []float64{1},
			wantErr: "stdin:3: line too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []float64
			err := readInput(strings.NewReader(tt.in), func(f float64) {
				got = append(got, f)
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("readInput() error = %v", err)
				}
			} else {
				var ie *inputError
				if !errors.As(err, &ie) {
					t.Fatalf("readInput() error = %v, want *inputError", err)
				}
				if diff := cmp.Diff(tt.wantErr, err.Error()); diff != "" {
					t.Errorf("error mismatch (-want +got):\n%s", diff)
				}
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("values mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReadInputScannerError(t *testing.T) {
	err := readInput(errReader{errors.New("boom")}, func(float64) {
		t.Fatal("process called")
	})
	var ie *inputError
	if !errors.As(err, &ie) {
		t.Fatalf("readInput() error = %v, want *inputError", err)
	}
	if diff := cmp.Diff("stdin:1: boom", err.Error()); diff != "" {
		t.Errorf("error mismatch (-want +got):\n%s", diff)
	}
}

func TestRunRejectsBadInputWithoutOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	outputCalled := false
	s := &st.St{
		Formatter: func(io.Writer, []string, [][]string) {
			outputCalled = true
		},
	}
	code := run(s, strings.NewReader("1\nx\n"), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if outputCalled {
		t.Fatal("table written after input error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if diff := cmp.Diff("stdin:2: invalid number \"x\"\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}
}

func TestRunTrimsAndSkipsBlanks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	s := &st.St{
		Formatter: st.PlainTextFormatter,
		Count:     true,
		Mean:      true,
	}
	code := run(s, strings.NewReader(" \n1\n\n 2 \n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "2") || !strings.Contains(out, "1.5") {
		t.Fatalf("stdout = %q, want count 2 and mean 1.5", out)
	}
	if strings.Contains(out, "NA") {
		t.Fatalf("stdout = %q, two values should not render NA", out)
	}
}

func TestRunNoValues(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "blank lines", in: "\n \n\t\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			outputCalled := false
			s := &st.St{
				Formatter: func(io.Writer, []string, [][]string) {
					outputCalled = true
				},
			}
			code := run(s, strings.NewReader(tt.in), &stdout, &stderr)
			if code != 1 {
				t.Fatalf("run() = %d, want 1", code)
			}
			if outputCalled {
				t.Fatal("table written for empty input")
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if diff := cmp.Diff("stdin: no values\n", stderr.String()); diff != "" {
				t.Errorf("stderr mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRunOneValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var header, row []string
	s := &st.St{
		Formatter: func(w io.Writer, h []string, data [][]string) {
			header = append([]string(nil), h...)
			row = append([]string(nil), data[0]...)
			io.WriteString(w, strings.Join(row, ","))
		},
		Count:    true,
		Min:      true,
		Max:      true,
		Sum:      true,
		Mean:     true,
		Stddev:   true,
		Variance: true,
	}
	code := run(s, strings.NewReader("5\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	check := func(want, got []string) {
		t.Helper()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
	check([]string{"N", "MIN", "MAX", "SUM", "MEAN", "STDDEV", "VARIANCE"}, header)
	check([]string{"1", "5", "5", "5", "5", "NA", "NA"}, row)
	if diff := cmp.Diff("1,5,5,5,5,NA,NA", stdout.String()); diff != "" {
		t.Errorf("stdout mismatch (-want +got):\n%s", diff)
	}
}

func TestRunOneValuePlainText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	s := &st.St{Formatter: st.PlainTextFormatter}
	code := run(s, strings.NewReader("5\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "NA") || strings.Contains(out, "NaN") {
		t.Fatalf("stdout = %q, want NA and not NaN", out)
	}
	for _, col := range []string{"N", "MIN", "MAX", "SUM", "MEAN", "STDDEV"} {
		if !strings.Contains(out, col) {
			t.Fatalf("stdout = %q, missing %s", out, col)
		}
	}
}

// errReader fails on the first read, so Scanner.Err must be checked.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }
