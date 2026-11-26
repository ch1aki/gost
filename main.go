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
	Format   string `short:"f" long:"format" description:"fmt format for float columns (default: fixed-point decimal)"`
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
	if code := cli(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func cli(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var opts options
	opts.Version = func() {
		fmt.Fprintln(stdout, "0.1.1")
		os.Exit(0)
	}
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "gost"
	parser.Usage = "[OPTIONS] [FILE...]"
	rest, err := parser.ParseArgs(args)
	if err != nil {
		var fe *flags.Error
		if errors.As(err, &fe) && fe.Type == flags.ErrHelp {
			return 0
		}
		return 1
	}
	if err := validateFloatFormat(opts.Format); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}

	f := st.PlainTextFormatter
	if opts.Markdown {
		f = st.MarkdownTableFormatter
	}

	s := st.St{
		Formatter: f,
		Format:    opts.Format,

		Count:    opts.Count,
		Min:      opts.Min,
		Max:      opts.Max,
		Sum:      opts.Sum,
		Mean:     opts.Mean,
		Stddev:   opts.Stddev,
		Variance: opts.Vatiance,
	}

	return run(&s, rest, stdin, stdout, stderr)
}

// validateFloatFormat reports whether format is a fmt verb string that
// consumes exactly one float64. An empty format selects the historical
// fixed-point default and is not passed to fmt.
func validateFloatFormat(format string) error {
	if format == "" {
		return nil
	}
	if strings.Contains(fmt.Sprintf(format, 0.0), "%!") {
		return fmt.Errorf("invalid format %q", format)
	}
	return nil
}

// run reads numbers from files into s and, on success, writes one table to stdout.
// No files, or "-", means stdin. Files are concatenated. Input errors are one
// stderr line and return 1 without a table.
func run(s *st.St, files []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(files) == 0 {
		files = []string{"-"}
	}
	n := 0
	for _, name := range files {
		rc, label, err := openInput(name, stdin)
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		err = readInput(label, rc, func(v float64) {
			n++
			s.Process(v)
		})
		rc.Close()
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
	}
	if n == 0 {
		fmt.Fprintln(stderr, noValuesMessage(files))
		return 1
	}
	s.Output(stdout)
	return 0
}

func noValuesMessage(files []string) string {
	names := make([]string, len(files))
	for i, name := range files {
		if name == "-" {
			names[i] = "stdin"
		} else {
			names[i] = name
		}
	}
	return strings.Join(names, ", ") + ": no values"
}

func openInput(name string, stdin io.Reader) (io.ReadCloser, string, error) {
	if name == "-" {
		return io.NopCloser(stdin), "stdin", nil
	}
	f, err := os.Open(name)
	if err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) {
			return nil, "", fmt.Errorf("%s: %s", name, pe.Err)
		}
		return nil, "", err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, "", fmt.Errorf("%s: %s", name, err)
	}
	if info.IsDir() {
		f.Close()
		return nil, "", fmt.Errorf("%s: is a directory", name)
	}
	return f, name, nil
}

// maxLineSize is the longest accepted line. A line may hold many numbers, so
// this is far above bufio.Scanner's 64KiB default. A longer line still fails
// with <file>:<line>: line too long.
const maxLineSize = 8 << 20

// inputError is a rejected record. Error text is "<name>:<line>: <reason>".
type inputError struct {
	name   string
	line   int
	reason string
}

func (e *inputError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.name, e.line, e.reason)
}

// readInput scans r for whitespace-separated numbers. Blank lines are skipped.
// Non-numeric, NaN, and Inf values, and lines longer than maxLineSize,
// return *inputError named with name and the line number.
func readInput(name string, r io.Reader, process func(float64)) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxLineSize)
	line := 0
	for scanner.Scan() {
		line++
		for _, text := range strings.Fields(scanner.Text()) {
			f, err := strconv.ParseFloat(text, 64)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return &inputError{name: name, line: line, reason: fmt.Sprintf("unsupported value %q", text)}
			}
			if err != nil {
				return &inputError{name: name, line: line, reason: fmt.Sprintf("invalid number %q", text)}
			}
			process(f)
		}
	}
	if err := scanner.Err(); err != nil {
		line++
		reason := err.Error()
		var pe *os.PathError
		if errors.As(err, &pe) {
			reason = pe.Err.Error()
		}
		if errors.Is(err, bufio.ErrTooLong) {
			reason = "line too long"
		}
		return &inputError{name: name, line: line, reason: reason}
	}
	return nil
}
