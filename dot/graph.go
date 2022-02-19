package dot

import (
	"fmt"
	"strings"

	"github.com/berquerant/gotypegraph/util"
)

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
