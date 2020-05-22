package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	st "github.com/ch1aki/gost/lib"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Version  func() `short:"v" long:"version" description:"show version"`
	Format   string `short:"f" long:"format" default:"%g" description:"outpuut format"`
	Count    bool   `short:"N" long:"count" description:"sample size"`
	Min      bool   `long:"min" description:"minimum"`
	Max      bool   `long:"max" description:"maximum"`
	Sum      bool   `long:"sum" description:"sum of elements of the sample"`
	Mean     bool   `short:"m" long:"mean" description:"mean"`
	Stddev   bool   `long:"sd" description:"standard deviation"`
	Vatiance bool   `long:"variance" description:"variance"`
}

func main() {
	var opts options

	opts.Version = func() {
		fmt.Println("0.0.4")
		os.Exit(0)
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	s := st.St{
		Formatter: st.PlainTextFormatter,

		Count:    opts.Count,
		Min:      opts.Min,
		Max:      opts.Max,
		Sum:      opts.Sum,
		Mean:     opts.Mean,
		Stddev:   opts.Stddev,
		Variance: opts.Vatiance,
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		f, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		s.Process(f)
	}

	s.Output(os.Stdout)
}
