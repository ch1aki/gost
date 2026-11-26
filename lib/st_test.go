package st

import (
	"bytes"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func checkDiff(t *testing.T, want, got any, opts ...cmp.Option) {
	t.Helper()
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestSetOption(t *testing.T) {
	w := new(bytes.Buffer)
	st := St{
		Formatter: func(writer io.Writer, header []string, data [][]string) {
			s := []byte(strings.Join(header, ","))
			writer.Write(s)
		},
		Sum: true,
	}
	st.Output(w)

	checkDiff(t, "SUM", w.String())
}
func TestDefaultOption(t *testing.T) {
	w := new(bytes.Buffer)
	st := St{
		Formatter: func(writer io.Writer, header []string, data [][]string) {
			s := []byte(strings.Join(header, ","))
			writer.Write(s)
		},
	}
	st.Output(w)

	checkDiff(t, "N,MIN,MAX,SUM,MEAN,STDDEV", w.String())
}

func TestNotDefaultOption(t *testing.T) {
	w := new(bytes.Buffer)
	st := St{
		Formatter: func(writer io.Writer, header []string, data [][]string) {
			s := []byte(strings.Join(header, ","))
			writer.Write(s)
		},
		Variance: true,
	}
	st.Output(w)

	checkDiff(t, "VARIANCE", w.String())
}

func TestPlainTextFromatter(t *testing.T) {
	stdout := new(bytes.Buffer)
	st := St{
		Formatter: PlainTextFormatter,

		Count:  true,
		Min:    true,
		Max:    true,
		Sum:    true,
		Mean:   true,
		Stddev: true,
	}

	for _, v := range []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		st.Process(v)
	}
	st.Output(stdout)

	expected := `N 	MIN	MAX	SUM	MEAN	STDDEV             
10	0  	9  	45 	4.5 	3.0276503540974917	
`
	checkDiff(t, expected, stdout.String())
}

func TestMarkdownTableFromatter(t *testing.T) {
	stdout := new(bytes.Buffer)
	st := St{
		Formatter: MarkdownTableFormatter,

		Count:  true,
		Min:    true,
		Max:    true,
		Sum:    true,
		Mean:   true,
		Stddev: true,
	}

	for _, v := range []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		st.Process(v)
	}
	st.Output(stdout)

	expected := `| N  | MIN | MAX | SUM | MEAN |       STDDEV       |
|----|-----|-----|-----|------|--------------------|
| 10 |   0 |   9 |  45 |  4.5 | 3.0276503540974917 |
`
	checkDiff(t, expected, stdout.String())
}

// moments is the sample accumulated by Process. Variance and Stddev use the
// sample definition (divide by n-1), which is what Process computes today.
// Fields are exported so go-cmp can report them; the type itself stays local.
type moments struct {
	Count    int64
	Min      float64
	Max      float64
	Sum      float64
	Mean     float64
	Variance float64
	Stddev   float64
}

func TestProcessMoments(t *testing.T) {
	seq := make([]float64, 100)
	for i := 1; i <= 100; i++ {
		seq[i-1] = float64(i)
	}
	// Sample variance of 1..n is n(n+1)/12. For n=100 that is 2525/3.
	seqVar := 100.0 * 101.0 / 12.0

	// n=0 is rejected by the CLI (no table). n=1 prints NA for the sample
	// moments; see TestFewSamples.
	tests := []struct {
		name  string
		input []float64
		want  moments
	}{
		{
			name:  "seq 1..100",
			input: seq,
			want: moments{
				Count:    100,
				Min:      1,
				Max:      100,
				Sum:      5050,
				Mean:     50.5,
				Variance: seqVar,
				Stddev:   math.Sqrt(seqVar),
			},
		},
		{
			name:  "negatives",
			input: []float64{-2, -1, 0},
			want: moments{
				Count:    3,
				Min:      -2,
				Max:      0,
				Sum:      -3,
				Mean:     -1,
				Variance: 1,
				Stddev:   1,
			},
		},
	}

	opt := cmp.Options{
		cmpopts.EquateApprox(0, 1e-12),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s St
			for _, v := range tt.input {
				s.Process(v)
			}
			got := moments{
				Count:    s.count,
				Min:      s.min,
				Max:      s.max,
				Sum:      s.sum,
				Mean:     s.mean,
				Variance: s.variance,
				Stddev:   s.stddev,
			}
			checkDiff(t, tt.want, got, opt)
		})
	}
}

func TestFewSamples(t *testing.T) {
	tests := []struct {
		name   string
		input  float64
		st     St
		header []string
		row    []string
	}{
		{
			name:   "n=1 defaults",
			input:  5,
			header: []string{"N", "MIN", "MAX", "SUM", "MEAN", "STDDEV"},
			row:    []string{"1", "5", "5", "5", "5", "NA"},
		},
		{
			name:   "n=1 with variance",
			input:  5,
			st:     St{Count: true, Min: true, Max: true, Sum: true, Mean: true, Stddev: true, Variance: true},
			header: []string{"N", "MIN", "MAX", "SUM", "MEAN", "STDDEV", "VARIANCE"},
			row:    []string{"1", "5", "5", "5", "5", "NA", "NA"},
		},
		{
			name:   "n=1 negative",
			input:  -4,
			st:     St{Min: true, Max: true, Sum: true, Mean: true, Stddev: true, Variance: true},
			header: []string{"MIN", "MAX", "SUM", "MEAN", "STDDEV", "VARIANCE"},
			row:    []string{"-4", "-4", "-4", "-4", "NA", "NA"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.st
			var header []string
			var row []string
			s.Formatter = func(_ io.Writer, h []string, data [][]string) {
				header = append([]string(nil), h...)
				row = append([]string(nil), data[0]...)
			}
			s.Process(tt.input)
			s.Output(io.Discard)
			checkDiff(t, tt.header, header)
			checkDiff(t, tt.row, row)
		})
	}
}

func TestDefaultFloatFormat(t *testing.T) {
	for _, v := range []float64{1234567, 50005000, 1e-7, 0.000000123} {
		s := St{Sum: true}
		s.Process(v)
		var row []string
		s.Formatter = func(_ io.Writer, _ []string, data [][]string) {
			row = append([]string(nil), data[0]...)
		}
		s.Output(io.Discard)
		want := strconv.FormatFloat(v, 'f', -1, 64)
		checkDiff(t, []string{want}, row)
		if strings.Contains(row[0], "e") || strings.Contains(row[0], "E") {
			t.Fatalf("default format of %v = %q, want fixed-point", v, row[0])
		}
	}
}

func TestExplicitGFormat(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"1234567", "1.234567e+06"},
		{"50005000", "5.0005e+07"},
		{"0.0000001", "1e-07"},
	} {
		v, err := strconv.ParseFloat(tc.in, 64)
		if err != nil {
			t.Fatal(err)
		}
		s := St{Format: "%g", Max: true}
		s.Process(v)
		var row []string
		s.Formatter = func(_ io.Writer, _ []string, data [][]string) {
			row = append([]string(nil), data[0]...)
		}
		s.Output(io.Discard)
		checkDiff(t, []string{tc.want}, row)
	}
}

func TestPopulationMoments(t *testing.T) {
	s := St{Population: true, Count: true, Mean: true, Stddev: true, Variance: true}
	s.Process(1)
	s.Process(3)
	var row []string
	s.Formatter = func(_ io.Writer, _ []string, data [][]string) {
		row = append([]string(nil), data[0]...)
	}
	s.Output(io.Discard)
	// Population variance of {1, 3} is 1, so stddev is 1. Sample would be 2 and sqrt(2).
	checkDiff(t, []string{"2", "2", "1", "1"}, row)

	one := St{Population: true, Stddev: true, Variance: true}
	one.Process(5)
	var oneRow []string
	one.Formatter = func(_ io.Writer, _ []string, data [][]string) {
		oneRow = append([]string(nil), data[0]...)
	}
	one.Output(io.Discard)
	checkDiff(t, []string{"0", "0"}, oneRow)
}

func TestFormatVerb(t *testing.T) {
	s := St{
		Format:   "%.2f",
		Count:    true,
		Mean:     true,
		Stddev:   true,
		Variance: true,
	}
	for _, v := range []float64{1, 2} {
		s.Process(v)
	}
	var header, row []string
	s.Formatter = func(_ io.Writer, h []string, data [][]string) {
		header = append([]string(nil), h...)
		row = append([]string(nil), data[0]...)
	}
	s.Output(io.Discard)
	checkDiff(t, []string{"N", "MEAN", "STDDEV", "VARIANCE"}, header)
	// N stays an integer. Sample variance of {1, 2} is 0.5.
	checkDiff(t, []string{"2", "1.50", "0.71", "0.50"}, row)
}
