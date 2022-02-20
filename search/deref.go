package search

import (
	"go/ast"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

type (
	ObjExtractor interface {
		Extract(pkg *packages.Package, ident *ast.Ident) (Object, bool)
	}

	objExtractor struct{}
)

func NewObjExtractor() ObjExtractor {
	return &objExtractor{}
}

func (*objExtractor) Extract(pkg *packages.Package, ident *ast.Ident) (Object, bool) {
	obj, ok := pkg.TypesInfo.Defs[ident]
	logger.Verbosef("[ObjExtractor] %s (%s) %s %#v", pkg.Name, pkg.ID, ident, obj)
	return obj, ok
}
