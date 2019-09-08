package st

import (
	"log"

	"github.com/montanaflynn/stats"
)

type Results struct {
	Count                                 int
	Min, Max, Sum, Mean, Stddev, Variance float64
}

func St(input []float64) (*Results, error) {

	count := len(input)

	min, err := stats.Min(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	max, err := stats.Max(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	sum, err := stats.Sum(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	mean, err := stats.Mean(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	stddev, err := stats.StdDevS(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	variance, err := stats.VarS(input)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Results{Count: count, Min: min, Max: max, Sum: sum, Mean: mean, Stddev: stddev, Variance: variance}, nil

}
