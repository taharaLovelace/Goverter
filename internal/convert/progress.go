package convert

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type Progress struct {
	Input   string
	Current int
	Total   int
	Percent float64
	Done    bool
}

type ProgressFunc func(Progress)

func ParseProgress(reader io.Reader, duration float64, update func(float64, bool)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var elapsed float64
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		switch key {
		case "out_time_us", "out_time_ms":
			microseconds, err := strconv.ParseFloat(value, 64)
			if err == nil {
				elapsed = microseconds / 1_000_000
			}
		case "progress":
			done := value == "end"
			percent := float64(0)
			if duration > 0 {
				percent = elapsed / duration * 100
				if percent > 100 {
					percent = 100
				}
			}
			if done {
				percent = 100
			}
			update(percent, done)
		}
	}
	return scanner.Err()
}
