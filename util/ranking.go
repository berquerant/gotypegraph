package util

import "sort"

type Ranking interface {
	Add(v int)
	GetPercentile(v int) float64
}

func NewRanking(v ...int) Ranking {
	s := &ranking{
		d: make(map[int]int),
	}
	for _, x := range v {
		s.Add(x)
	}
	return s
}

type ranking struct {
	d map[int]int
}

func (s *ranking) Add(v int) { s.d[v]++ }

func (s *ranking) sorted() []int {
	var (
		i      int
		values = make([]int, len(s.d))
	)
	for k := range s.d {
		values[i] = k
		i++
	}
	sort.Ints(values)
	return values
}

func (s *ranking) GetPercentile(v int) float64 {
	values := s.sorted()
	if v < values[0] {
		return 0
	}
	for i, x := range values {
		if x > v {
			return float64(i) / float64(len(values))
		}
	}
	return 1
}

type Percentiler interface {
	Percentile(p float64) int
}

func NewPercentiler(min, max int) Percentiler {
	return &percentiler{
		min: min,
		max: max,
	}
}

type percentiler struct {
	min int
	max int
}

func (s *percentiler) Percentile(p float64) int {
	switch {
	case p < 0:
		return s.min
	case p > 1:
		return s.max
	default:
		width := s.max - s.min
		val := int(p * float64(width))
		return s.min + val
	}
}
