package main

import (
	"bytes"
	"errors"
	"github.com/ajstarks/svgo/float"
	"honnef.co/go/js/dom"
	"strconv"
	"time"
)

func Chart(element dom.Element) error {
	div, ok := element.(*dom.HTMLDivElement)
	if !ok {
		return errors.New("element was not a DIV")
	}

	//domElem := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "svg")
	//svgHtml, ok := domElem.(dom.HTMLElement)
	//if !ok {
	//	return errors.New("created svg is not HTMLElement")
	//}
	//svgElem, ok := domElem.(dom.SVGElement)
	//if !ok {
	//	return errors.New("created svg is not SVGElement")
	//}
	//
	//svgHtml.Style().Set("width", "100%")
	//svgHtml.Style().Set("height", "100%")
	//svgHtml.Style().Set("border", "1px solid black")

	svgBytes := &bytes.Buffer{}
	svgobj := svg.New(svgBytes)
	//svgobj.Startunit(100, 100, "%")
	ratio := float64(dom.GetWindow().InnerWidth()) / float64(dom.GetWindow().InnerHeight())
	width := 100 * ratio
	svgobj.StartviewUnit(100, 100, "%", 0, 0, 100*ratio, 100)
	svgobj.Rect(1, 1, width-2, 98, "fill: none; stroke: black;")
	//for _, pt := range ToPoints(Data, width-6, 94) {
	//	svgobj.Circle(pt.x+3, pt.y+3, 0.5)
	//}
	path := &bytes.Buffer{}
	begin := true
	for _, pt := range ToPoints(MovingAverage(5), width-6, 94) {
		if begin {
			path.WriteRune('M')
			begin = false
		} else {
			path.WriteRune('L')
		}
		path.WriteString(
			strconv.FormatFloat(pt.x+3, 'f', -1, 64) +
				" " +
				strconv.FormatFloat(pt.y+3, 'f', -1, 64) +
				" ",
		)
	}
	svgobj.Path(path.String(), "fill: none; stroke: black; stroke-width: 0.1")
	svgobj.End()

	div.SetInnerHTML(svgBytes.String())

	return nil
}

func ToPoints(data map[time.Time]*Node, width float64, height float64) (ret []struct{ x, y float64 }) {
	dates := Dates(data)
	if len(dates) > 0 {
		minDate := dates[0].Unix()
		maxDate := dates[len(dates)-1].Unix()
		dateScale := width / float64(maxDate-minDate)
		minData := -data[dates[0]].Samples[0]
		maxData := -data[dates[0]].Samples[0]
		for _, d := range dates {
			for _, s := range data[d].Samples {
				s := -s
				if s < minData {
					minData = s
				}
				if s > maxData {
					maxData = s
				}
			}
		}
		dataScale := height / (maxData - minData)
		for _, d := range dates {
			for _, s := range data[d].Samples {
				s := -s
				ret = append(ret, struct{ x, y float64 }{
					x: float64(d.Unix()-minDate) * dateScale,
					y: (s - minData) * dataScale,
				})
			}
		}
	}
	return
}
