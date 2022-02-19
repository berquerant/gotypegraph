package dot_test

import (
	"testing"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/stretchr/testify/assert"
)

type mockNodeList struct {
	str string
}

func (*mockNodeList) Len() int                    { return 0 }
func (*mockNodeList) Add(_ dot.Node) dot.NodeList { return nil }
func (*mockNodeList) Slice() []dot.Node           { return nil }
func (s *mockNodeList) String() string            { return s.str }

func TestSubgraph(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, `subgraph g {
}`, dot.NewSubgraph("g", nil).String())
	})

	t.Run("with nodes", func(t *testing.T) {
		assert.Equal(t, `subgraph g {
nodeList
}`, dot.NewSubgraph("g", &mockNodeList{
			str: "nodeList",
		}).String())
	})

	t.Run("with nodes and cluster", func(t *testing.T) {
		assert.Equal(t, `subgraph cluster_g {
nodeList
}`, dot.NewSubgraph("g", &mockNodeList{
			str: "nodeList",
		}, dot.WithSubgraphCluster(true)).String())
	})

	t.Run("with nodes and attrs", func(t *testing.T) {
		attrList := &mockAttrList{
			str: "attrList",
		}
		assert.Equal(t, `subgraph g {
attrList
nodeList
}`, dot.NewSubgraph("g", &mockNodeList{
			str: "nodeList",
		}, dot.WithSubgraphAttrList(attrList)).String())
		assert.True(t, attrList.asStmt)
	})
}

type mockEdgeList struct {
	str string
}

func (*mockEdgeList) Len() int                      { return 0 }
func (s *mockEdgeList) Add(_ dot.Edge) dot.EdgeList { return s }
func (*mockEdgeList) Slice() []dot.Edge             { return nil }
func (s *mockEdgeList) String() string              { return s.str }

type mockSubgraphList struct {
	str string
}

func (*mockSubgraphList) Len() int                              { return 0 }
func (s *mockSubgraphList) Add(_ dot.Subgraph) dot.SubgraphList { return s }
func (*mockSubgraphList) Slice() []dot.Subgraph                 { return nil }
func (s *mockSubgraphList) String() string                      { return s.str }

func TestGraph(t *testing.T) {
	t.Run("with nodes", func(t *testing.T) {
		assert.Equal(t, `strict digraph g {
nodeList
}`, dot.NewGraph("g", &mockNodeList{
			str: "nodeList",
		}, nil).String())
	})

	t.Run("with nodes and edges", func(t *testing.T) {
		assert.Equal(t, `strict digraph g {
nodeList
edgeList
}`, dot.NewGraph("g", &mockNodeList{
			str: "nodeList",
		}, &mockEdgeList{
			str: "edgeList",
		}).String())
	})

	t.Run("with nodes, edges, and attrs", func(t *testing.T) {
		attrs := &mockAttrList{
			str: "attrList",
		}
		assert.Equal(t, `strict digraph g {
attrList
nodeList
edgeList
}`, dot.NewGraph("g", &mockNodeList{
			str: "nodeList",
		}, &mockEdgeList{
			str: "edgeList",
		}, dot.WithGraphAttrList(attrs)).String())
		assert.True(t, attrs.asStmt)
	})

	t.Run("with nodes, edges, attrs and subgraphs", func(t *testing.T) {
		attrs := &mockAttrList{
			str: "attrList",
		}
		assert.Equal(t, `strict digraph g {
attrList
subgraphList
nodeList
edgeList
}`, dot.NewGraph("g", &mockNodeList{
			str: "nodeList",
		}, &mockEdgeList{
			str: "edgeList",
		}, dot.WithGraphAttrList(attrs),
			dot.WithGraphSubgraphList(&mockSubgraphList{
				str: "subgraphList",
			}),
		).String())
		assert.True(t, attrs.asStmt)
	})
}
