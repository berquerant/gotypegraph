package search

import (
	"go/token"
	"go/types"
)

// Object is the set of the public methods of types.Object.
type Object interface {
	Parent() *types.Scope
	Pos() token.Pos
	Pkg() *types.Package
	Name() string
	Type() types.Type
	Exported() bool
	Id() string
}
