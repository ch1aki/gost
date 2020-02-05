package main

import (
	"fmt"
	"os"
	"strings"

	st "github.com/ch1aki/gost/lib"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Version func() `short:"v" long:"version" description:"show version"`
	Format  string `short:"f" long:"format" default:"%g" description:"outpuut format"`
}

func main() {
	var opts options

	opts.Version = func() {
		fmt.Println("0.0.3")
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

	template := strings.Replace("%d\t{}\t{}\t{}\t{}\t{}\n", "{}", opts.Format, -1)
	fmt.Printf("N\tmin\tmax\tsum\tmean\tstddev\n")
	fmt.Printf(template, r.Count, r.Min, r.Max, r.Sum, r.Mean, r.Stddev)

}
