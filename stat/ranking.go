package stat

import "sort"

// Ranking is a discrete distribution.
type Ranking interface {
	Add(v int)
	GetPercentile(v int) float64
}

func NewRanking() Ranking {
	return &ranking{
		d: map[int]int{},
	}
}

type ranking struct {
	d map[int]int
}

func (s *ranking) Add(v int) { s.d[v]++ }
func (s *ranking) sortedValues() []int {
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

func (s *ranking) countLower(v int) int {
	values := s.sortedValues()
	if v < values[0] {
		return 0
	}
	for i, x := range values {
		if x > v {
			return i
		}
	}
	return len(values)
}

func (s *ranking) GetPercentile(v int) float64 {
	idx := s.countLower(v)
	return float64(idx) / float64(len(s.d))
}

// Percentiler converts a percent point into the percentile on the uniform distribution.
type Percentiler interface {
	Get(p float64) int
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

func (s *percentiler) Get(p float64) int {
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
