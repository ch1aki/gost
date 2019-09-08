package st

import (
	"bufio"
	"io"
	"log"
	"strconv"
)

func StdinParser(stdin io.Reader) ([]float64, error) {
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
