package search

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

/* search references */

type (
	RefPkgSearcher interface {
		Search(pkg *packages.Package, pos token.Pos) (ast.Node, bool)
	}

	RefSearcher interface {
		// Search returns a node if the pos is contained in the node.
		Search(defSet DefSet, pos token.Pos) (node ast.Node, found bool)
	}
)

func NewRefPkgSearcher(searcher RefSearcher, defSets []DefSet) RefPkgSearcher {
	defs := make(map[string]DefSet)
	for _, defSet := range defSets {
		defs[defSet.Pkg().ID] = defSet
	}
	return &refPkgSearcher{
		defSets:  defs,
		searcher: searcher,
	}
}

type refPkgSearcher struct {
	defSets  map[string]DefSet // pkg id => def set
	searcher RefSearcher
}

func (s *refPkgSearcher) Search(pkg *packages.Package, pos token.Pos) (ast.Node, bool) {
	if defSet, ok := s.defSets[pkg.ID]; ok {
		return s.searcher.Search(defSet, pos)
	}
	return nil, false
}

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
