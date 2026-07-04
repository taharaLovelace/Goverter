package pdf

import (
	"math"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestParseLayout(t *testing.T) {
	t.Parallel()
	layout, err := ParseLayout("A4", "LANDSCAPE", "small")
	if err != nil {
		t.Fatal(err)
	}
	if layout.PageSize != PageSizeA4 ||
		layout.Orientation != OrientationLandscape ||
		layout.Margin != MarginSmall {
		t.Fatalf("unexpected layout: %#v", layout)
	}
}

func TestParseLayoutRejectsInvalidCombinations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pageSize, orientation, margin string
	}{
		{"legal", "portrait", "none"},
		{"a4", "auto", "none"},
		{"a4", "portrait", "medium"},
		{"fit", "portrait", "small"},
	}
	for _, test := range tests {
		if _, err := ParseLayout(test.pageSize, test.orientation, test.margin); err == nil {
			t.Errorf("ParseLayout(%q, %q, %q) succeeded", test.pageSize, test.orientation, test.margin)
		}
	}
}

func TestImportConfigurationUsesLandscapeAndMargin(t *testing.T) {
	t.Parallel()
	layout := Layout{
		PageSize:    PageSizeA4,
		Orientation: OrientationLandscape,
		Margin:      MarginLarge,
	}
	configuration, err := layout.ImportConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if !configuration.PageDim.Landscape() {
		t.Fatalf("page dimensions are not landscape: %#v", configuration.PageDim)
	}
	if configuration.Pos != types.Center {
		t.Fatalf("position = %v, want center", configuration.Pos)
	}
	if configuration.Scale >= 1 || configuration.Scale <= 0 {
		t.Fatalf("scale = %f, want a value between zero and one", configuration.Scale)
	}

	expected := (configuration.PageDim.Height - 2*layout.marginPoints()) / configuration.PageDim.Height
	if math.Abs(configuration.Scale-expected) > 0.0001 {
		t.Fatalf("scale = %f, want %f", configuration.Scale, expected)
	}
}

func TestFitConfigurationUsesFullImagePage(t *testing.T) {
	t.Parallel()
	layout := Layout{
		PageSize:    PageSizeFit,
		Orientation: OrientationPortrait,
		Margin:      MarginNone,
	}
	configuration, err := layout.ImportConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Pos != types.Full {
		t.Fatalf("position = %v, want full", configuration.Pos)
	}
}
