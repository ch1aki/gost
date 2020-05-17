package st

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPlainTextFromatter(t *testing.T) {
	stdout := new(bytes.Buffer)
	st := St{Formatter: PlainTextFormatter}

	for _, v := range []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		st.Process(v)
	}
	st.Output(stdout)

	expected := `N 	MIN	MAX	SUM	MEAN	STDDEV             
10	0  	9  	45 	4.5 	3.0276503540974917	
`
	if diff := cmp.Diff(expected, stdout.String()); diff != "" {
		fmt.Printf("expected != actual\n%s\n", diff)
	}
}
