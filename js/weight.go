//go:generate gopherjs build -v -m . -o weight-new.js

package main

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"strconv"
	"sync"
	"time"
)

const DateBase = 25569

func getInt(from map[string]interface{}, key string) (int, error) {
	stringVal, err := getString(from, key)
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(stringVal)
	if err != nil {
		return 0, errors.New("key " + key + " not convertible to int: " + err.Error())
	}
	return intVal, nil
}

func getFloat(from map[string]interface{}, key string) (float64, error) {
	stringVal, err := getString(from, key)
	if err != nil {
		return 0, err
	}
	floatVal, err := strconv.ParseFloat(stringVal, 64)
	if err != nil {
		return 0, errors.New("key " + key + " not convertible to float64: " + err.Error())
	}
	return floatVal, nil
}

func getString(from map[string]interface{}, key string) (string, error) {
	v, ok := from[key]
	if !ok {
		return "", errors.New("key " + key + " not in map")
	}
	stringVal, ok := v.(string)
	if !ok {
		return "", errors.New("key " + key + " was not string")
	}
	return stringVal, nil
}

func processEntry(entry interface{}, dates map[int]time.Time) error {
	entryMap, ok := entry.(map[string]interface{})
	if !ok {
		return errors.New("entry is not map[string]interface{}")
	}
	cell, ok := entryMap["gs$cell"].(map[string]interface{})
	if !ok {
		return errors.New("entry is missing gs$cell")
	}

	var row, col int
	var input string
	var number float64
	var err error
	row, err = getInt(cell, "row")
	if err == nil {
		col, err = getInt(cell, "col")
	}
	if err == nil {
		input, err = getString(cell, "inputValue")
	}
	if err == nil && (col == 1 || col == 2) {
		number, err = getFloat(cell, "numericValue")
	}
	if err != nil {
		return err
	}
	var date time.Time
	if col != 1 {
		date, ok = dates[row]
		if !ok {
			return errors.New("no existing date found for non-1 column")
		}
	}
	switch col {
	case 1:
		secs := (number-DateBase)*86400 + 7*3600
		dates[row] = time.Unix(int64(secs), 0)
	case 2:
		addMeasurement(date, number)
	case 6:
		addAnnotation(date, input)
	}
	return nil
}

func loadData(data map[string]interface{}) {
	feed, ok := data["feed"].(map[string]interface{})
	if !ok {
		println("failed loading data, no `feed` or wasn't map[string]interface{}")
		return
	}

	entries, ok := feed["entry"].([]interface{})
	if !ok {
		println("failed loading data, no `feed.entry` or wasn't []interface{}")
		return
	}

	dates := map[int]time.Time{}
	for entryNum, entry := range entries {
		if err := processEntry(entry, dates); err != nil {
			println("skipping entry", entryNum, ",", fmt.Sprintf("%#+v", entry), "because", err)
		}
	}
}

var chartMu sync.Mutex

func updateChart() {
	chartMu.Lock()
	defer chartMu.Unlock()
	Chart(dom.GetWindow().Document().GetElementByID("chart_container").(*dom.HTMLDivElement))
}

func init() {
	js.Global.Set("loadData", loadData)
}

func main() {
	dom.GetWindow().AddEventListener("hashchange", false, func(dom.Event) {
		go func() {
			updateChart()
		}()
	})
	dom.GetWindow().Document().AddEventListener("DOMContentLoaded", false, func(event dom.Event) {
		go func() {
			updateChart()
		}()
	})
	dom.GetWindow().AddEventListener("resize", false, func(_ dom.Event) {
		go func() {
			updateChart()
		}()
	})
}
