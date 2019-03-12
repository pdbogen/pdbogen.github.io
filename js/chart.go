package main

import (
	"errors"
	"honnef.co/go/js/dom"
	"strconv"
	"time"
)

const (
	green = "#055F50"
	gold  = "#977808"
	plum  = "#76064C"
)

type ChartArea struct {
	Width, Height                                        float64
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft float64
	MinX, MinY, MaxX, MaxY                               float64
}

func (c ChartArea) ScaleX(x float64) float64 {
	return (x-c.MinX)/(c.MaxX-c.MinX)*(c.Width-c.PaddingLeft-c.PaddingRight) + c.PaddingLeft
}

func (c ChartArea) ScaleY(y float64) float64 {
	frac := (y - c.MinY) / (c.MaxY - c.MinY)
	scaled := (1 - frac) * (c.Height - c.PaddingTop - c.PaddingBottom)
	return c.PaddingTop + scaled
}

func Chart(div *dom.HTMLDivElement) error {
	area := ChartArea{
		Width:         div.OffsetWidth(),
		Height:        div.OffsetHeight() - 4,
		PaddingTop:    20,
		PaddingBottom: 80,
		PaddingLeft:   40,
		PaddingRight:  40,
	}

	svgDomElement := dom.GetWindow().Document().GetElementByID("svg")
	if svgDomElement == nil {
		svgDomElement = dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "svg")
		svgDomElement.SetID("svg")
		div.AppendChild(svgDomElement)
	}
	for _, node := range svgDomElement.ChildNodes() {
		svgDomElement.RemoveChild(node)
	}
	svgElem, ok := svgDomElement.(dom.SVGElement)
	if !ok {
		return errors.New("created svg is not SVGElement")
	}

	svgDomElement.(dom.SVGElement).SetAttribute(
		"viewBox",
		"0 0 "+
			strconv.FormatFloat(area.Width, 'f', 2, 64)+" "+
			strconv.FormatFloat(area.Height, 'f', 2, 64),
	)

	daysStr := dom.GetWindow().Location().Hash
	if len(daysStr) > 0 && daysStr[0] == '#' {
		daysStr = daysStr[1:]
	}
	var days int
	if daysStr == "" {
		days = 365
	} else {
		var err error
		days, err = strconv.Atoi(daysStr)

		if err != nil {
			println("ignoring error parsing hash to number of days:", err.Error())
		}
	}

	data := &Data
	if days != 0 {
		data = data.Last(time.Duration(days) * 24 * time.Hour)
	}
	dates := data.Dates()
	if days == 0 {
		days = int(dates[len(dates)-1].Sub(dates[0]).Hours()) / 24
	}
	println(days, "days in range")

	area.MinX, area.MaxX, area.MinY, area.MaxY = data.Bounds()

	gridLines(svgElem, area, data)

	println("adding dots")
	for _, pt := range data.ToPoints(area) {
		svgElem.AppendChild(Fill(green, Title(pt.Date, Circle(pt.X, pt.Y, 2))))
	}

	fiveday := data.MovingAverage(5).DropZeroes()
	thirtyday := data.MovingAverage(30).DropZeroes()
	svgElem.AppendChild(Fill("none", Stroke(gold, Path(fiveday.ToPoints(area), 0, 0))))
	svgElem.AppendChild(Fill("none", Stroke(plum, Path(thirtyday.ToPoints(area), 0, 0))))

	addCursor(svgElem, area)

	svgElem.AddEventListener("mousemove", true, updateCursor(div, area, fiveday, thirtyday))

	return nil
}

func gridLines(svgElem dom.SVGElement, area ChartArea, data *DataSet) {
	dates := data.Dates()
	begin := dates[0]
	end := dates[len(dates)-1]
	days := end.Sub(begin).Hours() / 24

	for yIdx := 0; yIdx < 12; yIdx++ {
		y := area.MinY + (area.MaxY-area.MinY)*float64(yIdx)/12
		label := Text(
			0,
			area.ScaleY(y),
			strconv.FormatFloat(y, 'f', 1, 64))
		label.SetAttribute("textLength", strconv.FormatFloat(area.PaddingLeft, 'f', 1, 64))
		svgElem.AppendChild(label)
		svgElem.AppendChild(Fill("none", Stroke("#C0C0C0", Line(
			area.ScaleX(area.MinX),
			area.ScaleY(y),
			area.ScaleX(area.MaxX),
			area.ScaleY(y),
		))))
	}

	x := time.Date(begin.Year(), begin.Month(), 1, 0, 0, 0, 0, begin.Location())
	for {
		if x.After(end) {
			break
		}

		sx, sy := area.ScaleX(float64(x.Unix())), area.Height
		if sx >= area.PaddingLeft {
			var dateFormat string
			if days > 90 {
				dateFormat = "Jan 2006"
			} else {
				dateFormat = "2 Jan 2006"
			}
			label := Text(sx, sy, time.Unix(x.Unix(), 0).Format(dateFormat))
			label.SetAttribute("transform",
				"rotate(-90,"+
					strconv.FormatFloat(sx, 'f', 2, 64)+","+
					strconv.FormatFloat(sy, 'f', 2, 64)+")",
			)
			label.SetAttribute("textLength", strconv.FormatFloat(area.PaddingBottom, 'f', 2, 64))
			svgElem.AppendChild(label)

			svgElem.AppendChild(Fill("none", Stroke("#F0F0F0", Line(sx, area.ScaleY(area.MinY), sx,
				area.ScaleY(area.MaxY)))))
		}

		if days > 365 {
			x = time.Date(x.Year(), x.Month()+3, 1, 0, 0, 0, 0, x.Location())
		} else if days > 90 {
			x = time.Date(x.Year(), x.Month()+1, 1, 0, 0, 0, 0, x.Location())
		} else if days > 30 {
			x = time.Date(x.Year(), x.Month(), x.Day()+7, 0, 0, 0, 0, x.Location())
		} else {
			x = time.Date(x.Year(), x.Month(), x.Day()+1, 0, 0, 0, 0, x.Location())
		}
	}
}

func addCursor(svgElem dom.SVGElement, area ChartArea) {
	cursor := Fill("none", Stroke("#C0C0C0", Line(
		0, area.PaddingTop,
		0, area.Height-area.PaddingBottom,
	)))
	cursor.SetID("cursor")
	svgElem.AppendChild(cursor)
	cursorText := Text(0, area.PaddingTop, "")
	cursorText.SetID("cursor-text")
	cursorText.SetAttribute("dy", "1em")
	svgElem.AppendChild(cursorText)

	fiveDot := Fill(gold, Stroke("none", Circle(0, 0, 4)))
	fiveDot.SetID("fivedot")
	svgElem.AppendChild(fiveDot)

	thirtyDot := Fill(plum, Stroke("none", Circle(0, 0, 4)))
	thirtyDot.SetID("thirtydot")
	svgElem.AppendChild(thirtyDot)

}

func updateCursor(div *dom.HTMLDivElement, area ChartArea, fiveday *DataSet, thirtyday *DataSet) func(event dom.Event) {
	return func(event dom.Event) {
		fivedot := dom.GetWindow().Document().GetElementByID("fivedot")
		thirtydot := dom.GetWindow().Document().GetElementByID("thirtydot")

		mev, ok := event.(*dom.MouseEvent)
		if !ok {
			println("mouseover event not MouseEvent?")
			return
		}
		x := mev.Get("pageX").Float() - div.OffsetLeft() - area.PaddingLeft
		if x < 0 {
			x = 0
		}
		if x > area.Width-area.PaddingRight {
			x = area.Width - area.PaddingRight
		}
		dateUnix := x/(area.Width-area.PaddingLeft-area.PaddingRight)*(area.MaxX-area.MinX) + area.MinX

		cursor := dom.GetWindow().Document().GetElementByID("cursor")
		xcoord := strconv.FormatFloat(area.ScaleX(dateUnix), 'f', 2, 64)
		cursor.SetAttribute("x1", xcoord)
		cursor.SetAttribute("x2", xcoord)
		date := time.Unix(int64(dateUnix), 0)
		text := dom.GetWindow().Document().GetElementByID("cursor-text").(dom.SVGElement)
		if x/area.Width > 0.75 {
			text.SetAttribute("text-anchor", "end")
			text.SetAttribute("dx", "-1em")
		}
		if x/area.Width < 0.25 {
			text.SetAttribute("text-anchor", "beginning")
			text.SetAttribute("dx", "0")
		}
		fivedot.SetAttribute("cx", xcoord)
		fivedot.SetAttribute("cy", strconv.FormatFloat(area.ScaleY(fiveday.ValueAt(date)), 'f', 2, 64))
		thirtydot.SetAttribute("cx", xcoord)
		thirtydot.SetAttribute("cy", strconv.FormatFloat(area.ScaleY(thirtyday.ValueAt(date)), 'f', 2, 64))
		text.SetAttribute("x", xcoord)

		for _, e := range text.GetElementsByTagName("tspan") {
			text.RemoveChild(e)
		}

		text.AppendChild(Tspan(date.Format("2006-01-02")))
		fivespan := Tspan("5-Day: " + strconv.FormatFloat(fiveday.ValueAt(date), 'f', 2, 64))
		fivespan.SetAttribute("x", xcoord)
		fivespan.SetAttribute("dy", "1em")
		text.AppendChild(fivespan)

		thirtyspan := Tspan("30-Day: " + strconv.FormatFloat(thirtyday.ValueAt(date), 'f', 2, 64))
		thirtyspan.SetAttribute("x", xcoord)
		thirtyspan.SetAttribute("dy", "1em")
		text.AppendChild(thirtyspan)
	}
}

type Point struct {
	Date           string
	X, Y, Original float64
}

// Bounds returns the boundaries; the x axis is represented as unix time, the Y access in the native unit
func (d DataSet) Bounds() (minX, maxX, minY, maxY float64) {
	var curMinX, curMaxX int
	first := true
	for date, point := range d.Nodes {
		date := int(date.Unix())
		pt := point.Average()
		if date < curMinX || first {
			curMinX = date
		}
		if date > curMaxX || first {
			curMaxX = date
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

func (d DataSet) ToPoints(area ChartArea) (ret []Point) {
	dates := d.Dates()
	if len(dates) > 0 {
		for _, date := range dates {
			for _, s := range d.Nodes[date].Samples {
				ret = append(ret, Point{
					Date:     date.Format("2006-01-02"),
					X:        area.ScaleX(float64(date.Unix())),
					Y:        area.ScaleY(s),
					Original: s,
				})
			}
		}
	}
	return ret
}
