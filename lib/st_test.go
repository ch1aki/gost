package st

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

	if diff := cmp.Diff(w.String(), "SUM"); diff != "" {
		t.Errorf("expected != actual\n%s\n", diff)
	}
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

	if diff := cmp.Diff(w.String(), "N,MIN,MAX,SUM,MEAN,STDDEV"); diff != "" {
		t.Errorf("expected != actual\n%s\n", diff)
	}
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

	if diff := cmp.Diff(w.String(), "VARIANCE"); diff != "" {
		t.Errorf("expected != actual\n%s\n", diff)
	}
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
	if diff := cmp.Diff(expected, stdout.String()); diff != "" {
		t.Errorf("expected != actual\n%s\n", diff)
	}
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
	if diff := cmp.Diff(expected, stdout.String()); diff != "" {
		t.Errorf("expected != actual\n%s\n", diff)
	}
}
