package dot_test

import (
	"testing"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/stretchr/testify/assert"
)

type mockAttrList struct {
	str    string
	size   int
	asStmt bool
}

func (s *mockAttrList) Len() int                    { return s.size }
func (s *mockAttrList) Add(_ dot.Attr) dot.AttrList { return s }
func (*mockAttrList) Slice() []dot.Attr             { return nil }
func (s *mockAttrList) String(asStmt bool) string {
	s.asStmt = asStmt
	return s.str
}

func TestNode(t *testing.T) {
	t.Run("no attrs", func(t *testing.T) {
		assert.Equal(t, "id;", dot.NewNode("id").String())
	})

	t.Run("with attrs", func(t *testing.T) {
		attrList := &mockAttrList{
			str:  "attrs",
			size: 1,
		}
		assert.Equal(t, "id attrs;", dot.NewNode("id", dot.WithNodeAttrList(attrList)).String())
		assert.False(t, attrList.asStmt)
	})
}
