package jsonify

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/berquerant/gotypegraph/search"
)

type (
	Pkg struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	Pos struct {
		Pos      token.Pos `json:"pos"`
		Position string    `json:"position,omitempty"`
	}

	Ident struct {
		Name string `json:"name"`
		P    *Pos   `json:"p"`
	}

	Obj struct {
		Str  string `json:"str"`
		Recv string `json:"recv,omitempty"`
		Type string `json:"type"`
		P    *Pos   `json:"p"`
		Name string `json:"name"`
	}

	Ref struct {
		Pkg   *Pkg   `json:"pkg"`
		Ident *Ident `json:"ident"`
		Obj   *Obj   `json:"obj"`
	}

	Def struct {
		Pkg *Pkg `json:"pkg"`
		Obj *Obj `json:"obj"`
	}

	Use struct {
		Ref *Ref `json:"ref"`
		Def *Def `json:"def"`
	}
)

func NewUse(node search.Use) *Use {
	return &Use{
		Ref: newRef(node.Ref()),
		Def: newDef(node.Def()),
	}
}

func newDef(node search.DefNode) *Def {
	return &Def{
		Pkg: newPkg(node.Pkg()),
		Obj: newObj(node),
	}
}

func newRef(node search.RefNode) *Ref {
	return &Ref{
		Pkg:   newPkg(node.Pkg()),
		Ident: newIdent(node.Ident(), node.Pkg()),
		Obj:   newObj(node),
	}
}

func newObj(node search.Node) *Obj {
	return &Obj{
		Recv: node.RecvString(search.WithNodeRawRecv(true)),
		Type: node.Type().String(),
		Str:  types.ObjectString(node.Obj().(types.Object), nil),
		P:    newPos(node.Obj().Pos(), node.Pkg()),
		Name: node.Name(),
	}
}

func newIdent(ident *ast.Ident, pkg search.Pkg) *Ident {
	return &Ident{
		Name: ident.Name,
		P:    newPos(ident.Pos(), pkg),
	}
}

func newPos(pos token.Pos, pkg search.Pkg) *Pos {
	var p Pos
	p.Pos = pos
	if pk := pkg.Pkg(); pk != nil {
		p.Position = pk.Fset.Position(pos).String()
	}
	return &p
}

func newPkg(pkg search.Pkg) *Pkg {
	return &Pkg{
		Name: pkg.Name(),
		Path: pkg.Path(),
	}
}
