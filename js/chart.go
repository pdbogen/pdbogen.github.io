package main

import (
	"errors"
	"honnef.co/go/js/dom"
	"math"
	"strconv"
)

const (
	green = "#055F50"
	gold  = "#977808"
	plum  = "#76064C"
)

func Chart(div *dom.HTMLDivElement) error {
	svgDomElement := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "svg")
	svgElem, ok := svgDomElement.(dom.SVGElement)
	if !ok {
		return errors.New("created svg is not SVGElement")
	}

	svgElem.SetAttribute("width", "100%")
	svgElem.SetAttribute("height", "100%")

	//width := float64(dom.GetWindow().InnerWidth())
	//height := float64(dom.GetWindow().InnerHeight())
	width := div.OffsetWidth()
	height := div.OffsetHeight()
	svgElem.AppendChild(Fill("white", Rect(0, 0, width, height)))

	svgElem.SetAttribute(
		"viewBox",
		"0 0 "+
			strconv.FormatFloat(width, 'f', 2, 64)+" "+
			strconv.FormatFloat(height, 'f', 2, 64),
	)
	svgElem.SetAttribute("xmlns", "http://www.w3.org/2000/svg")

	//svgElem.AppendChild(Rect(1, 1, width-2, height-2))

	minx, maxx, miny, maxy := Data.Bounds()

	for _, pt := range Data.ToPoints(width-6, height-6, minx, maxx, miny, maxy) {
		svgElem.AppendChild(Fill(green, Title(pt.Date, Circle(pt.X+3, pt.Y+3, 2))))
	}

	svgElem.AppendChild(Fill("none", Stroke(gold, Path(
		Data.
			MovingAverage(5).
			DropZeroes().
			ToPoints(width-6, height-6, minx, maxx, miny, maxy),
		3, 3))))
	svgElem.AppendChild(Fill("none", Stroke(plum, Path(
		Data.
			MovingAverage(30).
			DropZeroes().
			ToPoints(width-6, height-6, minx, maxx, miny, maxy),
		3, 3))))

	for y := float64(int(miny)); y < maxy; y += math.Floor((maxy - miny) / 12) {
		//svgElem.AppendChild(TextRight(0, (float64(y + 3) - miny))/(maxy-miny)*height, "label"))
		scaledY := height - (y-miny)/(maxy-miny)*height
		svgElem.AppendChild(TextRight(
			0,
			scaledY+3,
			strconv.FormatFloat(y, 'f', 0, 64)))
		svgElem.AppendChild(Fill("none", Stroke("#C0C0C0", Path(
			[]Point{
				{X: 0, Y: scaledY},
				{X: width, Y: scaledY},
			}, 3, 3))))
	}

	cursor := Fill("none", Stroke("#C0C0C0", Path([]Point{}, 0, 0)))
	cursor.SetID("cursor")
	svgElem.AppendChild(cursor)
	cursorText := Text(0, 0, "")
	cursorText.SetID("cursor-text")
	svgElem.AppendChild(cursorText)

	div.SetInnerHTML("")
	div.AppendChild(svgElem)

	svgElem.AddEventListener("mousemove", true, func(event dom.Event) {
		mev, ok := event.(*dom.MouseEvent)
		if !ok {
			println("mouseover event not MouseEvent?")
			return
		}
		println(div.OffsetTop())
		x := mev.Get("pageX").Float() - div.OffsetLeft()
		dom.GetWindow().Document().GetElementByID("cursor").SetAttribute("d",
			"M "+strconv.FormatFloat(x, 'f', 2, 64)+" 0"+
				"L"+strconv.FormatFloat(x, 'f', 2, 64)+" "+
				strconv.FormatFloat(height, 'f', 2, 64))
		text := dom.GetWindow().Document().GetElementByID("cursor-text")
		text.SetAttribute("x", strconv.Itoa(mev.ClientX))
		text.SetInnerHTML(
			"ClientX: " + strconv.FormatFloat(x, 'f', 2, 64),
		)
	})

	return nil
}

type Point struct {
	Date           string
	X, Y, Original float64
}

// Bounds returns the boundaries; the x axis is represented as unix time, the Y access in the native unit
func (d DataSet) Bounds() (minX, maxX, minY, maxY float64) {
	var curMinX, curMaxX int64
	first := true
	for date, point := range d {
		pt := point.Average()
		if date.Unix() < curMinX || first {
			curMinX = date.Unix()
		}
		if date.Unix() > curMaxX || first {
			curMaxX = date.Unix()
		}
		if pt < minY || first {
			minY = pt
		}
		if pt > maxY || first {
			maxY = pt
		}
		first = false
	}
	return float64(curMinX), float64(curMaxX), minY, maxY
}

func (d DataSet) ToPoints(width float64, height float64, minX float64, maxX float64, minY float64, maxY float64) (ret []Point) {
	dates := d.Dates()
	if len(dates) > 0 {
		for _, date := range dates {
			for _, s := range d[date].Samples {
				ret = append(ret, Point{
					Date:     date.Format("2006-01-02"),
					X:        (float64(date.Unix()) - minX) / (maxX - minX) * width,
					Y:        height - (s-minY)/(maxY-minY)*height,
					Original: s,
				})
			}
		}
	}
	return ret
}
