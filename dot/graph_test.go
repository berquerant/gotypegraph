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
