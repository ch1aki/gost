package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {

	app := cli.NewApp()
	app.Name = "gost"
	app.Usage = "simple compute statistics"
	app.Version = "0.1.0"

	app.Action = func(c *cli.Context) error {
		i, err := stdinParser(os.Stdin)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		r, err := St(i)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		fmt.Printf("N\tmin\tmax\tsum\tmean\tstddev\n")
		fmt.Printf("%d\t%g\t%g\t%g\t%g\t%g\n", r.Count, r.Min, r.Max, r.Sum, r.Mean, r.Stddev)

		return nil
	}

	app.Run(os.Args)

}

func stdinParser(stdin io.Reader) ([]float64, error) {
	buf := make([]float64, 0)
	s := bufio.NewScanner(stdin)

	for s.Scan() {
		f, err := strconv.ParseFloat(s.Text(), 64)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		buf = append(buf, f)
	}

	if s.Err() != nil {
		log.Fatal(s.Err())
	}

	return buf, nil
}
