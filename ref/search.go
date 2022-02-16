package ref

import (
	"go/ast"
	"go/token"

	"github.com/berquerant/gotypegraph/def"
	"golang.org/x/tools/go/packages"
)

type Searcher interface {
	// Search returns a node that contains the pos in the pkg.
	Search(pkg *packages.Package, pos token.Pos) (ast.Node, bool)
	Set() LocalSearcherSet
}

func NewSearcher(set LocalSearcherSet) Searcher {
	return &searcher{
		set: set,
	}
}

type searcher struct {
	set LocalSearcherSet
}

func (s *searcher) Search(pkg *packages.Package, pos token.Pos) (ast.Node, bool) {
	q, ok := s.set.Get(pkg)
	if !ok {
		return nil, false
	}
	if r, ok := q.Search(pos); ok {
		return r, true
	}
	return nil, false
}

func (s *searcher) Set() LocalSearcherSet { return s.set }

// LocalSearcherSet manages a set of the LocalSearcher.
type LocalSearcherSet interface {
	Get(pkg *packages.Package) (LocalSearcher, bool)
}

func NewLocalSearcherSet(sets []*def.Set) LocalSearcherSet {
	searchers := make(map[string]LocalSearcher, len(sets))
	for _, s := range sets {
		searchers[s.Pkg.ID] = NewLocalSearcher(s)
	}
	return &localSearcherSet{
		searchers: searchers,
	}
}

type localSearcherSet struct {
	searchers map[string]LocalSearcher
}

func (s *localSearcherSet) Get(pkg *packages.Package) (LocalSearcher, bool) {
	x, ok := s.searchers[pkg.ID]
	return x, ok
}

func NewLocalSearcher(set *def.Set) LocalSearcher {
	return &localSearcher{
		set: set,
	}
}

// LocalSearcher searches the pos in the defs.
type LocalSearcher interface {
	// Search returns a node if the pos is contained in the node.
	Search(pos token.Pos) (ast.Node, bool)
	Set() *def.Set
}

type localSearcher struct {
	set *def.Set
}

func (s *localSearcher) Set() *def.Set { return s.set }

func (s *localSearcher) Search(pos token.Pos) (ast.Node, bool) {
	for _, d := range s.set.Defs {
		for _, vs := range d.ValueSpecs {
			if s.contains(vs, pos) {
				return vs, true
			}
		}
		for _, fd := range d.FuncDecls {
			if s.contains(fd, pos) {
				return fd, true
			}
		}
		for _, ts := range d.TypeSpecs {
			if s.contains(ts, pos) {
				return ts, true
			}
		}
	}
	return nil, false
}

func (s *localSearcher) contains(node ast.Node, pos token.Pos) bool {
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
