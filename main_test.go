package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
			if !strings.Contains(text, "[OPTIONS] [FILE...]") {
				t.Fatalf("help text = %q, want file arguments in Usage", text)
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
	long := strings.Repeat("1", maxLineSize+1) + "\n"

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
			err := readInput("stdin", strings.NewReader(tt.in), func(f float64) {
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
	err := readInput("stdin", errReader{errors.New("boom")}, func(float64) {
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
	code := run(s, nil, strings.NewReader("1\nx\n"), &stdout, &stderr)
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
	code := run(s, nil, strings.NewReader(" \n1\n\n 2 \n"), &stdout, &stderr)
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
			code := run(s, nil, strings.NewReader(tt.in), &stdout, &stderr)
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
	code := run(s, nil, strings.NewReader("5\n"), &stdout, &stderr)
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
	code := run(s, nil, strings.NewReader("5\n"), &stdout, &stderr)
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

func TestReadInputLongLineOfNumbers(t *testing.T) {
	var b strings.Builder
	const n = 20000
	for i := 1; i <= n; i++ {
		if i > 1 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%d", i)
	}
	b.WriteString(" x\n")
	var got []float64
	err := readInput("nums.txt", strings.NewReader(b.String()), func(f float64) {
		got = append(got, f)
	})
	var ie *inputError
	if !errors.As(err, &ie) {
		t.Fatalf("readInput() error = %v", err)
	}
	if diff := cmp.Diff("nums.txt:1: invalid number \"x\"", err.Error()); diff != "" {
		t.Errorf("error mismatch (-want +got):\n%s", diff)
	}
	if len(got) != n || got[0] != 1 || got[n-1] != n {
		t.Fatalf("processed %d values, first=%v last=%v", len(got), first(got), last(got))
	}
}

func first(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return v[0]
}

func last(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return v[len(v)-1]
}

func TestReadInputWhitespaceAndName(t *testing.T) {
	var got []float64
	err := readInput("data.txt", strings.NewReader("1 2\n\n\t3 x\n"), func(f float64) {
		got = append(got, f)
	})
	var ie *inputError
	if !errors.As(err, &ie) {
		t.Fatalf("readInput() error = %v", err)
	}
	if diff := cmp.Diff("data.txt:3: invalid number \"x\"", err.Error()); diff != "" {
		t.Errorf("error mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]float64{1, 2, 3}, got); diff != "" {
		t.Errorf("values mismatch (-want +got):\n%s", diff)
	}
}

func TestRunFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("1 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("\n3\t4\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	var header, row []string
	s := &st.St{
		Count: true,
		Sum:   true,
		Formatter: func(_ io.Writer, h []string, data [][]string) {
			header = append([]string(nil), h...)
			row = append([]string(nil), data[0]...)
		},
	}
	code := run(s, []string{a, "-", b}, strings.NewReader("5\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if diff := cmp.Diff([]string{"N", "SUM"}, header); diff != "" {
		t.Errorf("header mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"5", "15"}, row); diff != "" {
		t.Errorf("N and SUM mismatch (-want +got):\n%s", diff)
	}

	stdout.Reset()
	stderr.Reset()
	missing := filepath.Join(dir, "missing.txt")
	code = run(s, []string{missing}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("missing file code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if diff := cmp.Diff(missing+": no such file or directory\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}
	if strings.Contains(stderr.String(), "open ") {
		t.Fatalf("stderr repeats the open path: %q", stderr.String())
	}

	empty := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(empty, []byte(" \n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run(&st.St{Formatter: st.PlainTextFormatter}, []string{empty}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("empty file code=%d stdout=%q", code, stdout.String())
	}
	if diff := cmp.Diff(empty+": no values\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}

	empty2 := filepath.Join(dir, "empty2.txt")
	if err := os.WriteFile(empty2, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run(&st.St{Formatter: st.PlainTextFormatter}, []string{empty, empty2}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("empty files code=%d stdout=%q", code, stdout.String())
	}
	if diff := cmp.Diff(empty+", "+empty2+": no values\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}

	stdout.Reset()
	stderr.Reset()
	code = run(&st.St{Formatter: st.PlainTextFormatter}, []string{dir}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("directory code=%d stdout=%q", code, stdout.String())
	}
	if diff := cmp.Diff(dir+": is a directory\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}
}

func TestCLIPopulation(t *testing.T) {
	restore := captureOS(t)
	defer restore()
	var stdout, stderr bytes.Buffer
	code := cli([]string{"--population", "-m", "--sd", "--variance"}, strings.NewReader("1\n3\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("cli() = %d, stderr = %s", code, stderr.String())
	}
	var header, row []string
	probe := &st.St{
		Population: true,
		Formatter: func(_ io.Writer, h []string, data [][]string) {
			header = append([]string(nil), h...)
			row = append([]string(nil), data[0]...)
		},
		Mean:     true,
		Stddev:   true,
		Variance: true,
	}
	code = run(probe, nil, strings.NewReader("1\n3\n"), io.Discard, io.Discard)
	if code != 0 {
		t.Fatal(code)
	}
	if diff := cmp.Diff([]string{"MEAN", "STDDEV", "VARIANCE"}, header); diff != "" {
		t.Errorf("header mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"2", "1", "1"}, row); diff != "" {
		t.Errorf("population stddev mismatch (-want +got):\n%s\ncli stdout:\n%s", diff, stdout.String())
	}
	if strings.Contains(stdout.String(), "1.414") {
		t.Fatalf("cli stdout used the sample stddev:\n%s", stdout.String())
	}
}

func TestCLIFileArgs(t *testing.T) {
	restore := captureOS(t)
	defer restore()
	dir := t.TempDir()
	name := filepath.Join(dir, "nums.txt")
	if err := os.WriteFile(name, []byte("1 x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli([]string{"-N", name}, strings.NewReader("9\n"), &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 {
		t.Fatalf("cli() = %d, stdout = %q, stderr = %s", code, stdout.String(), stderr.String())
	}
	if diff := cmp.Diff(name+":1: invalid number \"x\"\n", stderr.String()); diff != "" {
		t.Errorf("stderr mismatch (-want +got):\n%s", diff)
	}
}

// errReader fails on the first read, so Scanner.Err must be checked.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }
