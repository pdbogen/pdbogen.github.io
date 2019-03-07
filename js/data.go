package main

import (
	"sort"
	"time"
)

type DataSet struct {
	Nodes   map[time.Time]*Node
	dates   []time.Time
	valueAt map[time.Time]float64
}

var Data = DataSet{}

type Node struct {
	Samples     []float64
	Annotations []string
}

func (n Node) ToPoint() {}

func (d *DataSet) Dates() (ret []time.Time) {
	if d.dates != nil {
		if len(d.Nodes) != len(d.dates) {
			panic("non-nil but mismatched dates")
		}
		return d.dates
	}
	ret = make([]time.Time, len(d.Nodes))
	i := 0
	for k := range d.Nodes {
		ret[i] = k
		i += 1
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Before(ret[j])
	})
	d.dates = ret
	return
}

func (n Node) Average() (ret float64) {
	for _, f := range n.Samples {
		ret += f / float64(len(n.Samples))
	}
	return
}

func (d *DataSet) nodeFor(date time.Time) *Node {
	t := date.Truncate(time.Hour * 24)
	node, ok := d.Nodes[t]
	if !ok {
		node = &Node{}
		if d.Nodes == nil {
			d.Nodes = map[time.Time]*Node{}
		}
		d.Nodes[t] = node
		d.dates = nil
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

func (d *DataSet) MovingAverage(days int) *DataSet {
	dates := d.Dates()
	begin := dates[0]
	end := dates[len(dates)-1]
	ret := &DataSet{Nodes: map[time.Time]*Node{}}
	for date := begin; date.Before(end.Add(time.Hour * 24)); date = date.Add(time.Hour * 24) {
		ret.Nodes[date] = &Node{}
		for i := -days + 1; i <= 0; i++ {
			if data, ok := d.Nodes[date.Add(time.Duration(i)*time.Hour*24)]; ok {
				ret.Nodes[date].Samples = append(ret.Nodes[date].Samples, data.Average())
			}
		}
		ret.Nodes[date].Samples = []float64{ret.Nodes[date].Average()}
	}
	_ = ret.Dates()
	return ret
}

func (d *DataSet) Last(duration time.Duration) *DataSet {
	ret := &DataSet{Nodes: map[time.Time]*Node{}}
	today := time.Now().Truncate(time.Hour * 24)
	for date := today.Add(-1 * duration).Truncate(time.Hour * 24); date.Before(today.Add(time.Hour * 24)); date = date.Add(time.Hour * 24) {
		if node, ok := d.Nodes[date]; ok {
			ret.Nodes[date] = node
		}
	}
	_ = ret.Dates()
	return ret
}

func (d *DataSet) DropZeroes() *DataSet {
	ret := &DataSet{Nodes: map[time.Time]*Node{}}
	for date, node := range d.Nodes {
		zero := true
		for _, s := range node.Samples {
			if s != 0 {
				zero = false
			}
		}
		if !zero {
			ret.Nodes[date] = node
		}
	}
	_ = ret.dates
	return ret
}

func (d *DataSet) ValueAt(t time.Time) float64 {
	if d.valueAt == nil {
		d.valueAt = map[time.Time]float64{}
	}

	//t = t.Truncate(time.Hour * 24)
	if v, ok := d.valueAt[t]; ok {
		return v
	}

	if v, ok := d.Nodes[t]; ok {
		d.valueAt[t] = v.Average()
		return d.valueAt[t]
	}

	dates := d.Dates()
	firstAfterIdx := sort.Search(len(dates), func(i int) bool {
		return dates[i].After(t)
	})
	lastBefore := dates[firstAfterIdx-1]
	firstAfter := dates[firstAfterIdx]
	v1 := d.Nodes[lastBefore].Average()
	v2 := d.Nodes[firstAfter].Average()
	span := float64(firstAfter.Sub(lastBefore))
	delta := float64(t.Sub(lastBefore))
	ret := delta/span*v2 + (span-delta)/span*v1
	if d.valueAt == nil {
		d.valueAt = map[time.Time]float64{}
	}
	d.valueAt[t] = ret
	return ret
}
