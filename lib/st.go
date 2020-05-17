package st

import (
	"io"
	"math"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type St struct {
	Formatter func(writer io.Writer, header []string, data [][]string)

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
	header := []string{"N", "MIN", "MAX", "SUM", "MEAN", "STDDEV"}
	data := [][]string{
		[]string{
			strconv.FormatInt(s.count, 10),
			strconv.FormatFloat(s.min, 'f', -1, 64),
			strconv.FormatFloat(s.max, 'f', -1, 64),
			strconv.FormatFloat(s.sum, 'f', -1, 64),
			strconv.FormatFloat(s.mean, 'f', -1, 64),
			strconv.FormatFloat(s.stddev, 'f', -1, 64),
		},
	}
	s.Formatter(writer, header, data)
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
