package search

import (
	"go/ast"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

/* search definitions */

type (
	// Def is a set of the top level definitions of a package.
	Def interface {
		ValueSpecs() []*ast.ValueSpec // var, const, field
		FuncDecls() []*ast.FuncDecl   // func, method
		TypeSpecs() []*ast.TypeSpec   // type alias, defined type
		// Iterate visits all specs and decls until given function returns false.
		Iterate(func(ast.Node) bool)
	}

	DefExtractor interface {
		Extract(f *ast.File) Def
	}

	DefSet interface {
		Pkg() *packages.Package
		Defs() []Def
	}

	DefSetExtractor interface {
		Extract(pkg *packages.Package) DefSet
	}
)

func NewDefExtractor() DefExtractor { return &defExtractor{} }

type defExtractor struct{}

func (*defExtractor) Extract(f *ast.File) Def {
	var (
		valueSpecs []*ast.ValueSpec
		funcDecls  []*ast.FuncDecl
		typeSpecs  []*ast.TypeSpec
	)
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					logger.Debugf("[DefExtractor] TypeSpec %s %d", spec.Name, spec.Pos())
					typeSpecs = append(typeSpecs, spec)
				case *ast.ValueSpec:
					logger.Debugf("[DefExtractor] ValueSpec %v %d", spec.Names, spec.Pos())
					valueSpecs = append(valueSpecs, spec)
				}
			}
		case *ast.FuncDecl:
			logger.Debugf("[DefExtractor] FuncDecl %s %d", decl.Name, decl.Pos())
			funcDecls = append(funcDecls, decl)
		}
	}
	return NewDef(valueSpecs, funcDecls, typeSpecs)
}

func NewDefSetExtractor(extractor DefExtractor) DefSetExtractor {
	return &defSetExtractor{
		extractor: extractor,
	}
}

type defSetExtractor struct {
	extractor DefExtractor
}

func (s *defSetExtractor) Extract(pkg *packages.Package) DefSet {
	defs := make([]Def, len(pkg.Syntax))
	for i, f := range pkg.Syntax {
		logger.Debugf("[DefSetExtractor] load %s (%s) %d", pkg.Name, pkg.ID, i)
		defs[i] = s.extractor.Extract(f)
	}
	logger.Verbosef("[DefSetExtractor] %s (%s) %d files loaded", pkg.Name, pkg.ID, len(defs))
	return NewDefSet(pkg, defs)
}

func NewDef(valueSpecs []*ast.ValueSpec, funcDecls []*ast.FuncDecl, typeSpecs []*ast.TypeSpec) Def {
	return &def{
		valueSpecs: valueSpecs,
		funcDecls:  funcDecls,
		typeSpecs:  typeSpecs,
	}
}

type def struct {
	valueSpecs []*ast.ValueSpec
	funcDecls  []*ast.FuncDecl
	typeSpecs  []*ast.TypeSpec
}

func (s *def) ValueSpecs() []*ast.ValueSpec { return s.valueSpecs }
func (s *def) FuncDecls() []*ast.FuncDecl   { return s.funcDecls }
func (s *def) TypeSpecs() []*ast.TypeSpec   { return s.typeSpecs }

func (s *def) Iterate(f func(ast.Node) bool) {
	for _, fd := range s.funcDecls {
		if !f(fd) {
			return
		}
	}
	for _, ts := range s.typeSpecs {
		if !f(ts) {
			return
		}
	}
	for _, vs := range s.valueSpecs {
		if !f(vs) {
			return
		}
	}
}

func NewDefSet(pkg *packages.Package, defs []Def) DefSet {
	return &defSet{
		pkg:  pkg,
		defs: defs,
	}
}

type defSet struct {
	pkg  *packages.Package
	defs []Def
}

func (s *defSet) Pkg() *packages.Package { return s.pkg }
func (s *defSet) Defs() []Def            { return s.defs }
