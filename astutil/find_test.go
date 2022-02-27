package astutil_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/berquerant/gotypegraph/astutil"
	"github.com/stretchr/testify/assert"
)

func TestFindValueSpecIndex(t *testing.T) {
	for _, tc := range []struct {
		title    string
		src      string
		pos      token.Pos
		want     int
		notFound bool
	}{
		{
			title: "var",
			src: `package testpkg
var v1 = 0`,
			pos:  21,
			want: 0,
		},
		{
			title: "vars[0]",
			src: `package testpkg
var v1, v2 = 1, 2`,
			pos:  21,
			want: 0,
		},
		{
			title: "vars[1]",
			src: `package testpkg
var v1, v2 = 1, 2`,
			pos:  25,
			want: 1,
		},
		{
			title: "vars[0] value",
			src: `package testpkg
var v1, v2 = 1, 2`,
			pos:  30,
			want: 0,
		},
		{
			title: "vars[1] value",
			src: `package testpkg
var v1, v2 = 1, 2`,
			pos:  33,
			want: 1,
		},
		{
			title: "vars type",
			src: `package testpkg
var v1, v2 int = 1, 2`,
			pos:  28,
			want: 0,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tc.src, 0)
			if !assert.Nil(t, err) {
				return
			}
			ast.Print(fset, f)
			got, ok := astutil.FindValueSpecIndex(f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec), tc.pos)
			assert.Equal(t, tc.notFound, !ok)
			assert.Equal(t, tc.want, got)
		})
	}
}
