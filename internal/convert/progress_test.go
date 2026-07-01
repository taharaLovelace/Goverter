package convert

import (
	"strings"
	"testing"
)

func TestParseProgress(t *testing.T) {
	t.Parallel()
	input := strings.NewReader(
		"frame=1\nout_time_us=500000\nprogress=continue\n" +
			"out_time_us=2000000\nprogress=end\n",
	)
	var updates []float64
	err := ParseProgress(input, 2, func(percent float64, _ bool) {
		updates = append(updates, percent)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 2 || updates[0] != 25 || updates[1] != 100 {
		t.Fatalf("updates = %#v, want [25 100]", updates)
	}
}
