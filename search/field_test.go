package search_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/berquerant/gotypegraph/logger"
	"github.com/berquerant/gotypegraph/search"
	"github.com/stretchr/testify/assert"
)

func TestFieldSearcher(t *testing.T) {
	logger.SetLevel(logger.Debug)
	const src = `package testpkg
type X struct {
  Int int
  String string
}
type Y struct {
  X
  Y string
  Float float64
}
type Z struct {}
var Int = 1`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if !assert.Nil(t, err) {
		return
	}
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	var conf types.Config
	_, err = conf.Check("testpkg", fset, []*ast.File{f}, &info)
	if !assert.Nil(t, err) {
		return
	}

	pkg := func() *types.Package {
		for _, obj := range info.Defs {
			if obj != nil {
				return obj.Pkg() // testpkg
			}
		}
		return nil
	}()
	if !assert.NotNil(t, pkg) {
		return
	}

	searcher := search.NewFieldSearcher([]*types.Package{pkg})
	for _, tc := range []struct {
		title    string
		pos      token.Pos
		want     string
		notFound bool
	}{
		{
			title: "X.Int",
			pos:   35,
			want:  "X",
		},
		{
			title: "X.String",
			pos:   45,
			want:  "X",
		},
		{
			title: "Y.X",
			pos:   79,
			want:  "Y",
		},
		{
			title: "Y.Y",
			pos:   83,
			want:  "Y",
		},
		{
			title: "Y.Float",
			pos:   94,
			want:  "Y",
		},
		{
			title:    "Z",
			pos:      115,
			notFound: true,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			got, found := searcher.Search(pkg, tc.pos)
			assert.Equal(t, !tc.notFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.want, got.Name())
		})
	}
}
