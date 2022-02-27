package util

import "regexp"

type RegexpPair interface {
	Accept() *regexp.Regexp
	Deny() *regexp.Regexp
	MatchString(string) bool
}

func NewRegexpPair(accept, deny *regexp.Regexp) RegexpPair {
	return &regexpPair{
		accept: accept,
		deny:   deny,
	}
}

type regexpPair struct {
	accept *regexp.Regexp
	deny   *regexp.Regexp
}

func (s *regexpPair) Accept() *regexp.Regexp { return s.accept }
func (s *regexpPair) Deny() *regexp.Regexp   { return s.deny }

func (s *regexpPair) MatchString(v string) bool {
	return (s.accept == nil || s.accept.MatchString(v)) && (s.deny == nil || !s.deny.MatchString(v))
}
