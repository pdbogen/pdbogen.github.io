package main

import (
	"honnef.co/go/js/dom"
	"strconv"
)

func Title(title string, e dom.SVGElement) dom.SVGElement {
	titleElem := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "title")
	titleElem.SetInnerHTML(title)
	e.AppendChild(titleElem)
	return e
}

func Stroke(color string, e dom.SVGElement) dom.SVGElement {
	e.SetAttribute("stroke", color)
	return e
}

func Fill(color string, e dom.SVGElement) dom.SVGElement {
	e.SetAttribute("fill", color)
	return e
}

func Rect(x float64, y float64, w float64, h float64) (dom.SVGElement) {
	rect, ok := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "rect").(dom.SVGElement)
	if !ok {
		panic("document.CreateElement(rect) did not return dom.SVGElement")
	}
	rect.SetAttribute("x", strconv.FormatFloat(x, 'f', -1, 64))
	rect.SetAttribute("y", strconv.FormatFloat(y, 'f', -1, 64))
	rect.SetAttribute("width", strconv.FormatFloat(w, 'f', -1, 64))
	rect.SetAttribute("height", strconv.FormatFloat(h, 'f', -1, 64))
	return rect
}

func Circle(cx float64, cy float64, r float64) (dom.SVGElement) {
	circle, ok := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "circle").(dom.SVGElement)
	if !ok {
		panic("document.CreateElement(circle) did not return dom.SVGElement")
	}
	circle.SetAttribute("cx", strconv.FormatFloat(cx, 'f', -1, 64))
	circle.SetAttribute("cy", strconv.FormatFloat(cy, 'f', -1, 64))
	circle.SetAttribute("r", strconv.FormatFloat(r, 'f', -1, 64))
	return circle
}

func Path(pts []Point, shiftX float64, shiftY float64) (dom.SVGElement) {
	var cmd string
	for _, pt := range pts {
		if len(cmd) == 0 {
			cmd = "M"
		} else {
			cmd = cmd + "L"
		}
		cmd = cmd +
			strconv.FormatFloat(pt.X+shiftX, 'f', 2, 64) + "," +
			strconv.FormatFloat(pt.Y+shiftY, 'f', 2, 64) + " "
	}

	path, ok := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "path").(dom.SVGElement)
	if !ok {
		panic("document.CreateElement(path) did not return dom.SVGElement")
	}
	path.SetAttribute("d", cmd)
	path.SetAttribute("style", "stroke-width: 1px;")
	return path
}

func Text(x float64, y float64, text string) dom.SVGElement {
	textElem, ok := dom.GetWindow().Document().CreateElementNS("http://www.w3.org/2000/svg", "text").(dom.SVGElement)
	if !ok {
		panic("document.CreateElement(text) did not return dom.SVGElement")
	}
	textElem.SetAttribute("x", strconv.FormatFloat(x, 'f', 2, 64))
	textElem.SetAttribute("y", strconv.FormatFloat(y, 'f', 2, 64))
	textElem.SetAttribute("dy", "1em")
	textElem.SetInnerHTML(text)
	return textElem
}

func TextRight(x float64, y float64, text string) dom.SVGElement {
	textElem := Text(x, y, text)
	textElem.SetAttribute("text-anchor", "end")
	return textElem
}
