package st

import (
	"io"
	"math"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type St struct {
	Formatter func(writer io.Writer, header []string, data [][]string)
	Count     bool
	Min       bool
	Max       bool
	Sum       bool
	Mean      bool
	Stddev    bool

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
	s.variance = s.sumsq / float64(s.count-1)

	s.stddev = math.Sqrt(s.variance)
}

func (s *St) Output(writer io.Writer) {
	header := []string{}
	data := []string{}

	if s.Count {
		header = append(header, "N")
		data = append(data, strconv.FormatInt(s.count, 10))
	}
	if s.Min {
		header = append(header, "MIN")
		data = append(data, strconv.FormatFloat(s.min, 'f', -1, 64))
	}
	if s.Max {
		header = append(header, "MAX")
		data = append(data, strconv.FormatFloat(s.max, 'f', -1, 64))
	}
	if s.Sum {
		header = append(header, "SUM")
		data = append(data, strconv.FormatFloat(s.sum, 'f', -1, 64))
	}
	if s.Mean {
		header = append(header, "MEAN")
		data = append(data, strconv.FormatFloat(s.mean, 'f', -1, 64))
	}
	if s.Stddev {
		header = append(header, "STDDEV")
		data = append(data, strconv.FormatFloat(s.stddev, 'f', -1, 64))
	}

	s.Formatter(writer, header, [][]string{data})
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
