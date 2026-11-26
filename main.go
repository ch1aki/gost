package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

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
	Markdown bool   `long:"markdown" description:"markdown table format"`
}

func main() {
	var opts options

	opts.Version = func() {
		fmt.Println("0.1.1")
		os.Exit(0)
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	f := st.PlainTextFormatter
	if opts.Markdown {
		f = st.MarkdownTableFormatter
	}

	s := st.St{
		Formatter: f,

		Count:    opts.Count,
		Min:      opts.Min,
		Max:      opts.Max,
		Sum:      opts.Sum,
		Mean:     opts.Mean,
		Stddev:   opts.Stddev,
		Variance: opts.Vatiance,
	}

	if code := run(&s, os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

// run reads numbers from r into s and, on success, writes the table to stdout.
// Input errors are written to stderr as a single line and return 1 without a table.
func run(s *st.St, r io.Reader, stdout, stderr io.Writer) int {
	n := 0
	err := readInput(r, func(v float64) {
		n++
		s.Process(v)
	})
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	if n == 0 {
		fmt.Fprintln(stderr, "stdin: no values")
		return 1
	}
	s.Output(stdout)
	return 0
}

// inputError is a rejected stdin record. Error text is "stdin:<line>: <reason>".
type inputError struct {
	line   int
	reason string
}

func (e *inputError) Error() string {
	return fmt.Sprintf("stdin:%d: %s", e.line, e.reason)
}

// readInput scans r as one number per line. Blank lines are skipped.
// Surrounding whitespace is ignored. Non-numeric, NaN, and Inf values,
// and lines longer than the scanner limit, return *inputError.
//
// Lines past bufio.MaxScanTokenSize (64KiB) are rejected instead of being
// dropped. A numeric token that long is not a useful float64, and growing
// the buffer without a limit would hide the same class of read failure.
func readInput(r io.Reader, process func(float64)) error {
	scanner := bufio.NewScanner(r)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		f, err := strconv.ParseFloat(text, 64)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return &inputError{line: line, reason: fmt.Sprintf("unsupported value %q", text)}
		}
		if err != nil {
			return &inputError{line: line, reason: fmt.Sprintf("invalid number %q", text)}
		}
		process(f)
	}
	if err := scanner.Err(); err != nil {
		line++
		if errors.Is(err, bufio.ErrTooLong) {
			return &inputError{line: line, reason: "line too long"}
		}
		return &inputError{line: line, reason: err.Error()}
	}
	return nil
}
