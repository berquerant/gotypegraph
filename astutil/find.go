package astutil

import "go/ast"

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
