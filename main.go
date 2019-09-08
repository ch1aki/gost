package main

import (
	"fmt"
	"os"

	st "github.com/ch1aki/gost/lib"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Version func() `short:"v" long:"version" description:"show version"`
}

func main() {
	var opts options

	opts.Version = func() {
		fmt.Println("0.0.2")
		os.Exit(0)
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	i, err := st.StdinParser(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	r, err := st.St(i)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("N\tmin\tmax\tsum\tmean\tstddev\n")
	fmt.Printf("%d\t%g\t%g\t%g\t%g\t%g\n", r.Count, r.Min, r.Max, r.Sum, r.Mean, r.Stddev)

}
