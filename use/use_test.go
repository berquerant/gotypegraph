package use_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/berquerant/gotypegraph/def"
	"github.com/berquerant/gotypegraph/use"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

type mockObj struct {
	pkg *types.Package
	pos token.Pos
}

func (*mockObj) Parent() *types.Scope  { return nil }
func (s *mockObj) Pos() token.Pos      { return s.pos }
func (s *mockObj) Pkg() *types.Package { return s.pkg }
func (*mockObj) Name() string          { return "" }
func (s *mockObj) Type() types.Type    { return nil }
func (*mockObj) Exported() bool        { return false }
func (*mockObj) Id() string            { return "" }
func (*mockObj) String() string        { return "" }

func TestDefFilter(t *testing.T) {
	const src = `package testpkg
func main() {
  const internalVar = "int"
  println("in main")
}
var va = 1
const (
  ca = 2
  cb = 3
)
type X struct{}
var d1, d2 = 1, 2
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if !assert.Nil(t, err) {
		return
	}
	ast.Print(fset, f)
	set := def.NewSetExtractor(def.NewExtactor()).Extract(&packages.Package{ // FIXME: not unit test
		ID:     "testpkgid",
		Name:   "testpkg",
		Syntax: []*ast.File{f},
	})
	filter := use.DefFilter([]*def.Set{set})

	for _, tc := range []struct {
		title string
		ident string
		pos   token.Pos
		want  bool
	}{
		{
			title: "internalVar",
			pos:   39,
		},
		{
			title: "va",
			pos:   86,
			want:  true,
		},
		{
			title: "ca",
			pos:   103,
			want:  true,
		},
		{
			title: "cb",
			pos:   112,
			want:  true,
		},
		{
			title: "X",
			pos:   126,
			want:  true,
		},
		{
			title: "main",
			pos:   22,
			want:  true,
		},
		{
			title: "d1",
			pos:   141,
			want:  true,
		},
		{
			title: "d2",
			pos:   145,
			want:  true,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			got := filter(nil, &mockObj{
				pkg: types.NewPackage("", "testpkg"),
				pos: tc.pos,
			})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFilter(t *testing.T) {
	var (
		tf use.Filter = func(_ *ast.Ident, _ use.Object) bool { return true }
		ff use.Filter = func(_ *ast.Ident, _ use.Object) bool { return false }
	)
	for _, tc := range []struct {
		title  string
		filter use.Filter
		want   bool
	}{
		{
			title:  "f and t or t",
			filter: ff.And(tf).Or(tf),
			want:   true,
		},
		{
			title:  "t and f or t",
			filter: tf.And(ff).Or(tf),
			want:   true,
		},
		{
			title:  "t and t or f",
			filter: tf.And(tf).Or(ff),
			want:   true,
		},
		{
			title:  "f or f",
			filter: ff.Or(ff),
		},
		{
			title:  "f or t",
			filter: ff.Or(tf),
			want:   true,
		},
		{
			title:  "t or f",
			filter: tf.Or(ff),
			want:   true,
		},
		{
			title:  "t or t",
			filter: tf.Or(tf),
			want:   true,
		},
		{
			title:  "f and f",
			filter: ff.And(ff),
		},
		{
			title:  "f and t",
			filter: ff.And(tf),
		},
		{
			title:  "t and f",
			filter: tf.And(ff),
		},
		{
			title:  "t and t",
			filter: tf.And(tf),
			want:   true,
		},
		{
			title:  "t",
			filter: tf,
			want:   true,
		},
		{
			title:  "f",
			filter: ff,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.filter(nil, nil))
		})
	}
}
