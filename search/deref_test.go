package search_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/berquerant/gotypegraph/astutil"
	"github.com/berquerant/gotypegraph/logger"
	"github.com/berquerant/gotypegraph/search"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

type objExtractorTestcase struct {
	title string
	src   string
	ident string
	want  func(*testing.T, search.Object)
}

func (s *objExtractorTestcase) test(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", s.src, 0)
	if !assert.Nil(t, err) {
		return
	}
	id, found := astutil.FindIdentFromFile(f, s.ident)
	if !assert.True(t, found) {
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
	pkg := &packages.Package{
		ID:        "pkgid",
		Name:      "testpkg",
		Syntax:    []*ast.File{f},
		TypesInfo: &info,
	}

	got, ok := search.NewObjExtractor().Extract(pkg, id)
	if !assert.True(t, ok) {
		return
	}
	s.want(t, got)
}

func TestObjExtractor(t *testing.T) {
	logger.SetLevel(logger.Debug)
	const src = `package testpkg
func f1() {}
type T1 struct{}
var v1 = 1
const (
  c1 = 2
  c2 = 3
)
var vx1, vx2 = 4, 5`
	for _, tc := range []objExtractorTestcase{
		{
			title: "func",
			src:   src,
			ident: "f1",
			want: func(t *testing.T, obj search.Object) {
				fn, ok := obj.(*types.Func)
				if !assert.True(t, ok) {
					return
				}
				assert.Equal(t, "f1", fn.Name())
			},
		},
		{
			title: "type",
			src:   src,
			ident: "T1",
			want: func(t *testing.T, obj search.Object) {
				tn, ok := obj.(*types.TypeName)
				if !assert.True(t, ok) {
					return
				}
				assert.Equal(t, "T1", tn.Name())
			},
		},
		{
			title: "var",
			src:   src,
			ident: "v1",
			want: func(t *testing.T, obj search.Object) {
				vr, ok := obj.(*types.Var)
				if !assert.True(t, ok) {
					return
				}
				assert.Equal(t, "v1", vr.Name())
			},
		},
		{
			title: "vars",
			src:   src,
			ident: "vx1",
			want: func(t *testing.T, obj search.Object) {
				vr, ok := obj.(*types.Var)
				if !assert.True(t, ok) {
					return
				}
				assert.Equal(t, "vx1", vr.Name())
			},
		},
	} {
		tc := tc
		t.Run(tc.title, tc.test)
	}
}
