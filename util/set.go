package util

type StringSet interface {
	Add(string) StringSet
	In(string) bool
}

func NewStringSet(seed ...string) StringSet {
	d := make(map[string]bool)
	for _, x := range seed {
		d[x] = true
	}
	return &stringSet{
		d: d,
	}
}

type stringSet struct {
	d map[string]bool
}

func (s *stringSet) Add(v string) StringSet {
	s.d[v] = true
	return s
}

func (s *stringSet) In(v string) bool { return s.d[v] }
