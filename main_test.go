package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	st "github.com/ch1aki/gost/lib"
	"github.com/google/go-cmp/cmp"
)

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
}

// errReader fails on the first read, so Scanner.Err must be checked.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }
