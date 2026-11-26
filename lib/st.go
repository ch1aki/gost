package st

import (
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type St struct {
	Formatter func(writer io.Writer, header []string, data [][]string)
	// Format is an explicit fmt verb for float columns. Empty keeps the
	// historical fixed-point formatting. N is always an integer.
	Format     string
	Count      bool
	Min        bool
	Max        bool
	Sum        bool
	Mean       bool
	Stddev     bool
	Variance   bool
	Population bool

	count    int64
	min      float64
	max      float64
	sum      float64
	mean     float64
	stddev   float64
	variance float64
	sumsq    float64
}

func (s *St) Process(input float64) {
	if s.count == 0 {
		s.min = input
		s.max = input
	}

	s.count++

	s.sum += input

	if s.min > input {
		s.min = input
	}

	if s.max < input {
		s.max = input
	}

	oldM := s.mean
	s.mean += (input - s.mean) / float64(s.count)

	s.sumsq += (input - s.mean) * (input - oldM)
	s.variance, s.stddev = s.moments()
}

// moments derives variance and standard deviation from the running sum of
// squares. The divisor is chosen here, so Population can change after Process.
func (s *St) moments() (variance, stddev float64) {
	denom := float64(s.count - 1)
	if s.Population {
		denom = float64(s.count)
	}
	if denom == 0 {
		return math.NaN(), math.NaN()
	}
	variance = s.sumsq / denom
	return variance, math.Sqrt(variance)
}

func (s *St) Output(writer io.Writer) {
	header := []string{}
	data := []string{}

	defaults := !(s.Count || s.Min || s.Max || s.Sum || s.Mean || s.Stddev || s.Variance)

	if defaults || s.Count {
		header = append(header, "N")
		data = append(data, strconv.FormatInt(s.count, 10))
	}
	if defaults || s.Min {
		header = append(header, "MIN")
		data = append(data, s.formatFloat(s.min))
	}
	if defaults || s.Max {
		header = append(header, "MAX")
		data = append(data, s.formatFloat(s.max))
	}
	if defaults || s.Sum {
		header = append(header, "SUM")
		data = append(data, s.formatFloat(s.sum))
	}
	if defaults || s.Mean {
		header = append(header, "MEAN")
		data = append(data, s.formatFloat(s.mean))
	}
	variance, stddev := s.moments()
	if defaults || s.Stddev {
		header = append(header, "STDDEV")
		data = append(data, s.formatSampleMoment(stddev))
	}
	if s.Variance {
		header = append(header, "VARIANCE")
		data = append(data, s.formatSampleMoment(variance))
	}

	s.Formatter(writer, header, [][]string{data})
}

// formatFloat prints v. With no Format, this matches
// strconv.FormatFloat(v, 'f', -1, 64). An explicit Format is a fmt verb.
func (s *St) formatFloat(v float64) string {
	if s.Format == "" {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf(s.Format, v)
}

// formatSampleMoment renders variance or standard deviation.
// Sample mode is undefined for n < 2, so those cells are NA.
// Population mode divides by n, so a single value is 0 rather than NA.
func (s *St) formatSampleMoment(v float64) string {
	if !s.Population && s.count < 2 {
		return "NA"
	}
	return s.formatFloat(v)
}

func PlainTextFormatter(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}

func MarkdownTableFormatter(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}
