package main

import (
	"sort"
	"time"
)

type DataSet map[time.Time]*Node

var Data = DataSet{}

type Node struct {
	Samples     []float64
	Annotations []string
}

func (n Node) ToPoint() {}

func (d DataSet) Dates() (ret []time.Time) {
	ret = make([]time.Time, len(d))
	i := 0
	for k := range d {
		ret[i] = k
		i += 1
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Before(ret[j])
	})
	return
}

func (n Node) Average() (ret float64) {
	for _, f := range n.Samples {
		ret += f / float64(len(n.Samples))
	}
	return
}

func (d DataSet) nodeFor(date time.Time) *Node {
	t := date.Truncate(time.Hour * 24)
	node, ok := d[t]
	if !ok {
		node = &Node{}
		d[t] = node
	}
	return node
}

func addMeasurement(date time.Time, measurement float64) {
	day := Data.nodeFor(date)
	day.Samples = append(day.Samples, measurement)
	return
}

func addAnnotation(date time.Time, annotation string) {
	day := Data.nodeFor(date)
	day.Annotations = append(day.Annotations, annotation)
}

func (d DataSet) MovingAverage(days int) (ret DataSet) {
	dates := d.Dates()
	begin := dates[0]
	end := dates[len(dates)-1]
	ret = DataSet{}
	for date := begin.Add(time.Hour * 24); date.Before(end.Add(time.Hour * 24)); date = date.Add(time.Hour * 24) {
		ret[date] = &Node{}
		for i := -time.Duration(days) * time.Hour * 24; i <= 0; i += time.Hour * 24 {
			if data, ok := d[date.Add(i)]; ok {
				ret[date].Samples = append(ret[date].Samples, data.Average())
			}
		}
		ret[date].Samples = []float64{ret[date].Average()}
	}
	return
}

func (d DataSet) Last(duration time.Duration) DataSet {
	ret := DataSet{}
	today := time.Now().Truncate(time.Hour * 24)
	for date := today.Add(-1 * duration).Truncate(time.Hour * 24); date.Before(today.Add(time.Hour * 24)); date = date.Add(time.Hour * 24) {
		if node, ok := d[date]; ok {
			ret[date] = node
		}
	}
	return ret
}

func (d DataSet) DropZeroes() DataSet {
dates:
	for date, node := range d {
		for _, s := range node.Samples {
			if s != 0 {
				continue dates
			}
		}
		delete(d, date)
	}
	return d
}
