package dot_test

import (
	"testing"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/stretchr/testify/assert"
)

func TestEdge(t *testing.T) {
	t.Run("no attrs", func(t *testing.T) {
		assert.Equal(t, "nid1 -> nid2;", dot.NewEdge("nid1", "nid2").String())
	})

	t.Run("with attrs", func(t *testing.T) {
		assert.Equal(t, "nid1 -> nid2 attrs;", dot.NewEdge("nid1", "nid2", dot.WithEdgeAttrList(&mockAttrList{
			str:  "attrs",
			size: 1,
		})).String())
	})
}
