// Package dot provides an interface to generate dot lang statements.
package dot

import (
	"fmt"
	"strings"

	"github.com/berquerant/gotypegraph/util"
)

type Attr struct {
	Key   string
	Value string
	AsRaw bool
}

func (s *Attr) String() string {
	if s.AsRaw {
		return fmt.Sprintf("%s=%s", s.Key, s.Value)
	}
	return fmt.Sprintf(`%s="%s"`, s.Key, s.Value)
}
func (s *Attr) Statement() string {
	if s.AsRaw {
		return fmt.Sprintf("%s=%s;", s.Key, s.Value)
	}
	return fmt.Sprintf(`%s="%s";`, s.Key, s.Value)
}

type AttrOption func(*Attr)

func WithAsRaw(v bool) AttrOption {
	return func(a *Attr) {
		a.AsRaw = v
	}
}

type AttrList struct {
	List []*Attr
}

func NewAttrList() *AttrList {
	return &AttrList{
		List: []*Attr{},
	}
}

func (s *AttrList) Len() int { return len(s.List) }

func (s *AttrList) Add(key, value string, opt ...AttrOption) *AttrList {
	a := &Attr{
		Key:   key,
		Value: value,
	}
	for _, x := range opt {
		x(a)
	}
	s.List = append(s.List, a)
	return s
}

func (s *AttrList) String() string {
	ss := make([]string, len(s.List))
	for i, a := range s.List {
		ss[i] = a.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(ss, ","))
}

func (s *AttrList) StatementList() string {
	ss := make([]string, len(s.List))
	for i, a := range s.List {
		ss[i] = a.Statement()
	}
	return strings.Join(ss, "\n")
}

type Node struct {
	ID    string
	Attrs *AttrList
}

func NewNode(id string) *Node {
	return &Node{
		ID:    id,
		Attrs: NewAttrList(),
	}
}

func (s *Node) String() string {
	if s.Attrs.Len() == 0 {
		return fmt.Sprintf("%s;", s.ID)
	}
	return fmt.Sprintf(`%s %s;`, s.ID, s.Attrs)
}

type Edge struct {
	Src   string
	Dst   string
	Attrs *AttrList
}

func NewEdge(src, dst string) *Edge {
	return &Edge{
		Src:   src,
		Dst:   dst,
		Attrs: NewAttrList(),
	}
}

func (s *Edge) String() string {
	if s.Attrs.Len() == 0 {
		return fmt.Sprintf("%s -> %s;", s.Src, s.Dst)
	}
	return fmt.Sprintf("%s -> %s %s;", s.Src, s.Dst, s.Attrs)
}

type NodeList struct {
	List []*Node
}

func NewNodeList() *NodeList {
	return &NodeList{
		List: []*Node{},
	}
}

func (s *NodeList) Add(node *Node) *NodeList {
	s.List = append(s.List, node)
	return s
}

func (s *NodeList) Len() int { return len(s.List) }

func (s *NodeList) String() string {
	var b util.StringBuilder
	for _, n := range s.List {
		b.Writeln(n.String())
	}
	return b.String()
}

type EdgeList struct {
	List []*Edge
}

func NewEdgeList() *EdgeList {
	return &EdgeList{
		List: []*Edge{},
	}
}

func (s *EdgeList) Add(edge *Edge) *EdgeList {
	s.List = append(s.List, edge)
	return s
}

func (s *EdgeList) Len() int { return len(s.List) }

func (s *EdgeList) String() string {
	var b util.StringBuilder
	for _, e := range s.List {
		b.Writeln(e.String())
	}
	return b.String()
}

type Subgraph struct {
	ID        string
	IsCluster bool
	Attrs     *AttrList
	Nodes     *NodeList
}

func NewSubgraph(id string, isCluster bool) *Subgraph {
	return &Subgraph{
		ID:        id,
		IsCluster: isCluster,
		Attrs:     NewAttrList(),
		Nodes:     NewNodeList(),
	}
}

func (s *Subgraph) id() string {
	if s.IsCluster {
		return fmt.Sprintf("cluster_%s", s.ID)
	}
	return s.ID
}

func (s *Subgraph) String() string {
	var b util.StringBuilder
	b.Writelnf("subgraph %s {", s.id())
	if s.Attrs.Len() > 0 {
		b.Writeln(s.Attrs.StatementList())
	}
	if s.Nodes.Len() > 0 {
		b.Writeln(s.Nodes.String())
	}
	b.Writeln("}")
	return b.String()
}

type SubgraphList struct {
	List []*Subgraph
}

func NewSubgraphList() *SubgraphList {
	return &SubgraphList{
		List: []*Subgraph{},
	}
}

func (s *SubgraphList) Add(g *Subgraph) *SubgraphList {
	s.List = append(s.List, g)
	return s
}

func (s *SubgraphList) Len() int { return len(s.List) }

func (s *SubgraphList) String() string {
	var b util.StringBuilder
	for _, g := range s.List {
		b.Writeln(g.String())
	}
	return b.String()
}

type Graph struct {
	ID        string
	Subgraphs *SubgraphList
	Nodes     *NodeList
	Edges     *EdgeList
	Attrs     *AttrList
}

func NewGraph(id string) *Graph {
	return &Graph{
		ID:        id,
		Subgraphs: NewSubgraphList(),
		Nodes:     NewNodeList(),
		Edges:     NewEdgeList(),
		Attrs:     NewAttrList(),
	}
}

func (s *Graph) String() string {
	var b util.StringBuilder
	b.Writelnf("digraph %s {", s.ID)
	if s.Attrs.Len() > 0 {
		b.Writeln(s.Attrs.StatementList())
	}
	if s.Subgraphs.Len() > 0 {
		b.Writeln(s.Subgraphs.String())
	}
	if s.Nodes.Len() > 0 {
		b.Writeln(s.Nodes.String())
	}
	if s.Edges.Len() > 0 {
		b.Writeln(s.Edges.String())
	}
	b.Writeln("}")
	return b.String()
}
