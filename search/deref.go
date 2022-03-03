package search

import (
	"go/ast"
	"go/types"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

type (
	ObjExtractor interface {
		// Extract extracts an object with the ident.
		Extract(pkg *packages.Package, ident *ast.Ident) (Object, bool)
	}

	objExtractor struct{}
)

func NewObjExtractor() ObjExtractor {
	return &objExtractor{}
}

func (*objExtractor) Extract(pkg *packages.Package, ident *ast.Ident) (Object, bool) {
	obj, ok := pkg.TypesInfo.Defs[ident]
	logger.Debugf("[ObjExtractor] %s", types.ObjectString(obj, nil))
	return obj, ok
}
