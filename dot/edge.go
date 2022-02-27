package dot

import (
	"fmt"
	"strings"
)

type (
	Edge interface {
		From() ID
		To() ID
		AttrList() AttrList
		String() string
	}

	EdgeConfig struct {
		attrList AttrList
	}

	EdgeOption func(*EdgeConfig)

	edge struct {
		from ID
		to   ID
		conf *EdgeConfig
	}
)

func NewEdge(from, to ID, opt ...EdgeOption) Edge {
	var config EdgeConfig
	for _, x := range opt {
		x(&config)
	}
	return &edge{
		from: from,
		to:   to,
		conf: &config,
	}
}

func WithEdgeAttrList(attrList AttrList) EdgeOption {
	return func(c *EdgeConfig) {
		c.attrList = attrList
	}
}

func (s *edge) From() ID           { return s.from }
func (s *edge) To() ID             { return s.to }
func (s *edge) AttrList() AttrList { return s.conf.attrList }
func (s *edge) String() string {
	if s.conf.attrList == nil || s.conf.attrList.Len() == 0 {
		return fmt.Sprintf("%s -> %s;", s.from, s.to)
	}
	return fmt.Sprintf("%s -> %s %s;", s.from, s.to, s.conf.attrList.String(false))
}

type (
	EdgeList interface {
		Len() int
		Add(Edge) EdgeList
		Slice() []Edge
		String() string
	}

	edgeList struct {
		list []Edge
	}
)

func NewEdgeList() EdgeList {
	return &edgeList{}
}

func (s *edgeList) Len() int {
	if s == nil {
		return 0
	}
	return len(s.list)
}

func (s *edgeList) Slice() []Edge { return s.list }

func (s *edgeList) Add(edge Edge) EdgeList {
	s.list = append(s.list, edge)
	return s
}

func (s *edgeList) String() string {
	if len(s.list) == 0 {
		return ""
	}
	ss := make([]string, len(s.list))
	for i, e := range s.list {
		ss[i] = e.String()
	}
	return strings.Join(ss, "\n")
}
