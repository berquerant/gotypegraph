package astutil

import (
	"go/ast"
	"go/token"
)

func FindIdentFromFile(f *ast.File, identName string) (ident *ast.Ident, found bool) {
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if id.Name == identName {
				ident = id
				found = true
				return false
			}
		}
		return true
	})
	return
}

func FindValueSpecIndex(vs *ast.ValueSpec, pos token.Pos) (int, bool) {
	if vs.Type != nil {
		if vs.Type.Pos() <= pos && pos <= vs.Type.End() {
			return 0, true
		}
	}
	for i, nm := range vs.Names {
		if nm.Pos() <= pos && pos <= nm.End() {
			return i, true
		}
	}
	for i, v := range vs.Values {
		if v.Pos() <= pos && pos <= v.End() {
			return i, true
		}
	}
	return 0, false
}
