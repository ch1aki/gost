package main

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestSt(t *testing.T) {
	expected := &Results{
		Count:    10,
		Min:      0,
		Max:      9,
		Sum:      45,
		Mean:     4.5,
		Stddev:   3.0276503540974917,
		Variance: 9.166666666666666,
	}
	input := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	actual, err := St(input)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		fmt.Printf("expected != actual\n%s\n", diff)
	}
}
