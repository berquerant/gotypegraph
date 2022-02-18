package search_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/berquerant/gotypegraph/astutil"
	"github.com/berquerant/gotypegraph/search"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

type refSearcherTestcase struct {
	title     string
	src       string
	ident     string
	wantFound bool
	wantIdent string
}

func (s *refSearcherTestcase) test(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", s.src, 0)
	if !assert.Nil(t, err) {
		return
	}
	ast.Print(fset, f)
	set := search.NewDefSetExtractor(search.NewDefExtractor()).Extract(&packages.Package{ // FIXME: not unit test
		ID:     "testpkg",
		Syntax: []*ast.File{f},
	})
	ident, ok := astutil.FindIdentFromFile(f, s.ident)
	if assert.True(t, ok, "ident should be found") {
		return
	}
	t.Logf("ident %s pos %s", s.ident, fset.Position(ident.Pos()))
	got, ok := search.NewRefSearcher().Search(set, ident.Pos())
	assert.Equal(t, s.wantFound, ok)
	if !s.wantFound {
		return
	}
	switch got := got.(type) {
	case *ast.FuncDecl:
		assert.Equal(t, s.wantIdent, got.Name.Name, "funcDecl")
	case *ast.TypeSpec:
		assert.Equal(t, s.wantIdent, got.Name.Name, "typeSpec")
	case *ast.ValueSpec:
		for _, n := range got.Names {
			if s.wantIdent == n.Name {
				return
			}
		}
		t.Errorf("%s %+v valueSpec", s.wantIdent, got.Names)
	default:
		t.Errorf("unexpected result %#v", got)
	}
}

func TestRefSearcher(t *testing.T) {
	for _, tc := range []refSearcherTestcase{
		{
			title: "found in nested func decl",
			src: `package p
func main() {
  inner := func() {
    println("hello")
  }
  inner()
}
`,
			ident:     "println",
			wantFound: true,
			wantIdent: "main",
		},
		{
			title: "found in value spec",
			src: `package p
var height = func() int {
  return 1 + 2 * base
}()
`,
			ident:     "base",
			wantFound: true,
			wantIdent: "height",
		},
		{
			title: "found in type spec",
			src: `package p
type Rock struct {
  height Meter
}
`,
			ident:     "Meter",
			wantFound: true,
			wantIdent: "Rock",
		},
		{
			title: "found in func decl",
			src: `package p
func main() {
  println("hello")
}
`,
			ident:     "println",
			wantFound: true,
			wantIdent: "main",
		},
		{
			title: "not found because ident is at the top level value",
			src: `package p
var Hole = "穴"
`,
			ident: "Hole",
		},
		{
			title: "not found because ident is at the top level type",
			src: `package p
type Delete struct {}
`,
			ident: "Delete",
		},
		{
			title: "not found because ident is at the top level func",
			src: `package p
func main() {}
`,
			ident: "main",
		},
	} {
		tc := tc
		t.Run(tc.title, tc.test)
	}
}
