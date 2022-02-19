package dot

import (
	"fmt"
	"strings"

	"github.com/berquerant/gotypegraph/util"
)

type (
	Graph interface {
		ID() ID
		SubgraphList() SubgraphList
		NodeList() NodeList
		EdgeList() EdgeList
		AttrList() AttrList
		String() string
	}

	GraphConfig struct {
		attrList     AttrList
		subgraphList SubgraphList
	}

	GraphOption func(*GraphConfig)

	graph struct {
		id       ID
		nodeList NodeList
		edgeList EdgeList
		conf     *GraphConfig
	}
)

func NewGraph(id ID, nodeList NodeList, edgeList EdgeList, opt ...GraphOption) Graph {
	var config GraphConfig
	for _, x := range opt {
		x(&config)
	}
	return &graph{
		id:       id,
		nodeList: nodeList,
		edgeList: edgeList,
		conf:     &config,
	}
}

func WithGraphAttrList(attrList AttrList) GraphOption {
	return func(c *GraphConfig) {
		c.attrList = attrList
	}
}

func WithGraphSubgraphList(subgraphList SubgraphList) GraphOption {
	return func(c *GraphConfig) {
		c.subgraphList = subgraphList
	}
}

func (s *graph) ID() ID                     { return s.id }
func (s *graph) SubgraphList() SubgraphList { return s.conf.subgraphList }
func (s *graph) NodeList() NodeList         { return s.nodeList }
func (s *graph) EdgeList() EdgeList         { return s.edgeList }
func (s *graph) AttrList() AttrList         { return s.conf.attrList }
func (s *graph) String() string {
	var b util.StringBuilder
	b.Writelnf("strict digraph %s {", s.id)
	if s.conf.attrList != nil {
		b.Writelnf(s.conf.attrList.String(true))
	}
	if s.conf.subgraphList != nil {
		b.Writeln(s.conf.subgraphList.String())
	}
	if s.nodeList != nil {
		b.Writeln(s.nodeList.String())
	}
	if s.edgeList != nil {
		b.Writeln(s.edgeList.String())
	}
	b.Write("}")
	return b.String()
}

type (
	Subgraph interface {
		ID() ID
		AttrList() AttrList
		NodeList() NodeList
		String() string
	}

	SubgraphConfig struct {
		isCluster bool
		attrList  AttrList
	}

	SubgraphOption func(*SubgraphConfig)

	subgraph struct {
		id       ID
		nodeList NodeList
		conf     *SubgraphConfig
	}
)

func NewSubgraph(id ID, nodeList NodeList, opt ...SubgraphOption) Subgraph {
	var config SubgraphConfig
	for _, x := range opt {
		x(&config)
	}
	return &subgraph{
		id:       id,
		nodeList: nodeList,
		conf:     &config,
	}
}

func WithSubgraphAttrList(attrList AttrList) SubgraphOption {
	return func(c *SubgraphConfig) {
		c.attrList = attrList
	}
}

func WithSubgraphCluster(v bool) SubgraphOption {
	return func(c *SubgraphConfig) {
		c.isCluster = v
	}
}

func (s *subgraph) ID() ID             { return s.id }
func (s *subgraph) AttrList() AttrList { return s.conf.attrList }
func (s *subgraph) NodeList() NodeList { return s.nodeList }

func (s *subgraph) getID() string {
	if s.conf.isCluster {
		return fmt.Sprintf("cluster_%s", s.id)
	}
	return s.id.String()
}

func (s *subgraph) String() string {
	var b util.StringBuilder
	b.Writelnf("subgraph %s {", s.getID())
	if s.conf.attrList != nil {
		b.Writeln(s.conf.attrList.String(true))
	}
	if s.nodeList != nil {
		b.Writeln(s.nodeList.String())
	}
	b.Write("}")
	return b.String()
}

type (
	SubgraphList interface {
		Len() int
		Add(Subgraph) SubgraphList
		Slice() []Subgraph
		String() string
	}

	subgraphList struct {
		list []Subgraph
	}
)

func (s *subgraphList) Len() int {
	if s == nil {
		return 0
	}
	return len(s.list)
}

func (s *subgraphList) Slice() []Subgraph {
	if s == nil {
		return nil
	}
	return s.list
}

func (s *subgraphList) Add(g Subgraph) SubgraphList {
	s.list = append(s.list, g)
	return s
}

func (s *subgraphList) String() string {
	ss := make([]string, len(s.list))
	for i, g := range s.list {
		ss[i] = g.String()
	}
	return strings.Join(ss, "\n")
}
