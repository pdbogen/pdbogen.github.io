package main

import (
	"sort"
	"time"
)

var Data = map[time.Time]*Node{}

type Node struct {
	Samples     []float64
	Annotations []string
}

func Dates(data ...map[time.Time]*Node) (ret []time.Time) {
	if data == nil || len(data) == 0 {
		data = []map[time.Time]*Node{Data}
	}

	subj := data[0]

	ret = make([]time.Time, len(subj))
	i := 0
	for k := range subj {
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

func nodeFor(date time.Time) *Node {
	t := date.Truncate(time.Hour * 24)
	node, ok := Data[t]
	if !ok {
		node = &Node{}
		Data[t] = node
	}
	return node
}

func addMeasurement(date time.Time, measurement float64) {
	day := nodeFor(date)
	day.Samples = append(day.Samples, measurement)
	return
}

func addAnnotation(date time.Time, annotation string) {
	day := nodeFor(date)
	day.Annotations = append(day.Annotations, annotation)
}

func MovingAverage(days int) (ret map[time.Time]*Node) {
	dates := Dates()
	begin := dates[0]
	end := dates[len(dates)-1]
	ret = map[time.Time]*Node{}
	for d := begin.Add(time.Hour * 24); d.Before(end.Add(time.Hour * 24)); d = d.Add(time.Hour * 24) {
		ret[d] = &Node{}
		for i := 0; i < days; i++ {
			data, ok := Data[d.Add(time.Hour*24*time.Duration(-i))]
			if ok {
				ret[d].Samples = append(ret[d].Samples, data.Average())
			}
		}
	}
	return
}
