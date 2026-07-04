package pdf

import (
	"fmt"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type PageSize string

const (
	PageSizeA4     PageSize = "a4"
	PageSizeLetter PageSize = "letter"
	PageSizeFit    PageSize = "fit"
)

type Orientation string

const (
	OrientationPortrait  Orientation = "portrait"
	OrientationLandscape Orientation = "landscape"
)

type Margin string

const (
	MarginNone  Margin = "none"
	MarginSmall Margin = "small"
	MarginLarge Margin = "large"
)

type Layout struct {
	PageSize    PageSize    `json:"page_size"`
	Orientation Orientation `json:"orientation"`
	Margin      Margin      `json:"margin"`
}

func ParseLayout(pageSize, orientation, margin string) (Layout, error) {
	layout := Layout{
		PageSize:    PageSize(strings.ToLower(strings.TrimSpace(pageSize))),
		Orientation: Orientation(strings.ToLower(strings.TrimSpace(orientation))),
		Margin:      Margin(strings.ToLower(strings.TrimSpace(margin))),
	}
	switch layout.PageSize {
	case PageSizeA4, PageSizeLetter, PageSizeFit:
	default:
		return Layout{}, fmt.Errorf("unsupported page size %q; use a4, letter, or fit", pageSize)
	}
	switch layout.Orientation {
	case OrientationPortrait, OrientationLandscape:
	default:
		return Layout{}, fmt.Errorf("unsupported orientation %q; use portrait or landscape", orientation)
	}
	switch layout.Margin {
	case MarginNone, MarginSmall, MarginLarge:
	default:
		return Layout{}, fmt.Errorf("unsupported margin %q; use none, small, or large", margin)
	}
	if layout.PageSize == PageSizeFit && layout.Margin != MarginNone {
		return Layout{}, fmt.Errorf("margins require page size a4 or letter")
	}
	return layout, nil
}

func (l Layout) ImportConfiguration() (*pdfcpu.Import, error) {
	if l.PageSize == PageSizeFit {
		return pdfcpuapi.Import("position:full", types.POINTS)
	}

	form := "A4"
	if l.PageSize == PageSizeLetter {
		form = "Letter"
	}
	if l.Orientation == OrientationLandscape {
		form += "L"
	}

	dim, _, err := types.ParsePageFormat(form)
	if err != nil {
		return nil, err
	}
	margin := l.marginPoints()
	scale := min(
		(dim.Width-2*margin)/dim.Width,
		(dim.Height-2*margin)/dim.Height,
	)
	if scale <= 0 {
		return nil, fmt.Errorf("margin is too large for page size %s", l.PageSize)
	}

	description := fmt.Sprintf(
		"formsize:%s, position:c, scalefactor:%.6f rel",
		form,
		scale,
	)
	return pdfcpuapi.Import(description, types.POINTS)
}

func (l Layout) marginPoints() float64 {
	const pointsPerMillimeter = 72.0 / 25.4
	switch l.Margin {
	case MarginSmall:
		return 10 * pointsPerMillimeter
	case MarginLarge:
		return 20 * pointsPerMillimeter
	default:
		return 0
	}
}

type Options struct {
	Output      string
	PageSize    string
	Orientation string
	Margin      string
	Recursive   bool
	Overwrite   bool
}

type Summary struct {
	Output      string      `json:"output"`
	PageCount   int         `json:"page_count"`
	Images      []string    `json:"images"`
	PageSize    PageSize    `json:"page_size"`
	Orientation Orientation `json:"orientation"`
	Margin      Margin      `json:"margin"`
}
