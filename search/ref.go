package search

import (
	"go/ast"
	"go/token"
)

/* search references */

type (
	RefSearcher interface {
		// Search returns a node if the pos is contained in the node.
		Search(defSet DefSet, pos token.Pos) (node ast.Node, found bool)
	}
)

func NewRefSearcher() RefSearcher {
	return &refSearcher{}
}

type refSearcher struct{}

func (s *refSearcher) Search(defSet DefSet, pos token.Pos) (node ast.Node, found bool) {
	for _, defs := range defSet.Defs() {
		if found {
			break
		}
		defs.Iterate(func(n ast.Node) bool {
			if s.contains(n, pos) {
				node = n
				found = true
				return false
			}
			return true
		})
	}
	return
}

func (*refSearcher) contains(node ast.Node, pos token.Pos) bool {
	switch node := node.(type) {
	case *ast.FuncDecl:
		return node.Type.Params.Opening < pos && pos < node.Body.Rbrace
	case *ast.TypeSpec:
		return node.Type.Pos() <= pos && pos <= node.Type.End()
	case *ast.ValueSpec:
		return node.Type != nil && node.Type.Pos() <= pos && pos <= node.Type.End() ||
			len(node.Values) > 0 && node.Values[0].Pos() <= pos && pos <= node.Values[len(node.Values)-1].End()
	default:
		return false
	}
}
