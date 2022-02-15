package def_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/berquerant/gotypegraph/def"
	"github.com/stretchr/testify/assert"
)

type extractTestcase struct {
	title          string
	src            string
	wantValueSpecs [][]string
	wantFuncDecls  []string
	wantTypeSpecs  []string
}

func (s *extractTestcase) test(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", s.src, 0)
	assert.Nil(t, err)
	ast.Print(fset, f)
	got := def.NewExtactor().Extract(f)
	s.assertValueSpecs(t, got)
	s.assertFuncDecls(t, got)
	s.assertTypeSpecs(t, got)
}

func (s *extractTestcase) assertTypeSpecs(t *testing.T, got *def.Def) {
	assert.Equal(t, len(s.wantTypeSpecs), len(got.TypeSpecs))
	if len(got.TypeSpecs) == 0 {
		return
	}
	names := make([]string, len(got.TypeSpecs))
	for i, ts := range got.TypeSpecs {
		names[i] = ts.Name.String()
	}
	assert.Equal(t, s.wantTypeSpecs, names)
}

func (s *extractTestcase) assertFuncDecls(t *testing.T, got *def.Def) {
	assert.Equal(t, len(s.wantFuncDecls), len(got.FuncDecls))
	if len(got.FuncDecls) == 0 {
		return
	}
	names := make([]string, len(got.FuncDecls))
	for i, fd := range got.FuncDecls {
		names[i] = fd.Name.String()
	}
	assert.Equal(t, s.wantFuncDecls, names)
}

func (s *extractTestcase) assertValueSpecs(t *testing.T, got *def.Def) {
	assert.Equal(t, len(s.wantValueSpecs), len(got.ValueSpecs))
	if len(got.ValueSpecs) == 0 {
		return
	}
	names := make([][]string, len(got.ValueSpecs))
	for i, vs := range got.ValueSpecs {
		nms := make([]string, len(vs.Names))
		for j, n := range vs.Names {
			nms[j] = n.String()
		}
		names[i] = nms
	}
	assert.Equal(t, s.wantValueSpecs, names)
}

func TestExtractor(t *testing.T) {
	for _, tc := range []*extractTestcase{
		{
			title: "a func",
			src: `package p
func main() {
  println("hello")
}`,
			wantFuncDecls: []string{
				"main",
			},
		},
		{
			title: "a type",
			src: `package p
type Yen int`,
			wantTypeSpecs: []string{
				"Yen",
			},
		},
		{
			title: "a var",
			src: `package p
var Global = "north"`,
			wantValueSpecs: [][]string{
				{
					"Global",
				},
			},
		},
		{
			title: "funcs",
			src: `package p
func F1(_ string) bool { return true }
func F2(_ int) string { return "true" }`,
			wantFuncDecls: []string{
				"F1",
				"F2",
			},
		},
		{
			title: "types",
			src: `package p
type planet string
type Secondary = uintptr
`,
			wantTypeSpecs: []string{
				"planet",
				"Secondary",
			},
		},
		{
			title: "vars",
			src: `package p
var V1 = 1
var V2, V3 = 2, 3
const C1 = 0`,
			wantValueSpecs: [][]string{
				{
					"V1",
				},
				{
					"V2", "V3",
				},
				{
					"C1",
				},
			},
		},
		{
			title: "a method",
			src: `package p
type Empty struct{}
func (*Empty) String() string {return "null"}`,
			wantTypeSpecs: []string{
				"Empty",
			},
			wantFuncDecls: []string{
				"String",
			},
		},
	} {
		tc := tc
		t.Run(tc.title, tc.test)
	}
}
