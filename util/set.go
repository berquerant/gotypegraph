package util

type StringSet struct {
	d map[string]bool
}

func NewStringSet() *StringSet {
	return &StringSet{
		d: map[string]bool{},
	}
}

func (s *StringSet) Add(v string) *StringSet {
	s.d[v] = true
	return s
}
func (s *StringSet) In(v string) bool { return s.d[v] }
func (s *StringSet) Len() int         { return len(s.d) }
func (s *StringSet) Slice() []string {
	var (
		i int
		r = make([]string, len(s.d))
	)
	for k := range s.d {
		r[i] = k
		i++
	}
	return r
}
