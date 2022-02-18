package search

import (
	"go/ast"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

/* search definitions */

type (
	Def interface {
		ValueSpecs() []*ast.ValueSpec // var, const, field
		FuncDecls() []*ast.FuncDecl   // func, method
		TypeSpecs() []*ast.TypeSpec   // type alias, defined type
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
					typeSpecs = append(typeSpecs, spec)
				case *ast.ValueSpec:
					valueSpecs = append(valueSpecs, spec)
				}
			}
		case *ast.FuncDecl:
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
		logger.Debugf("[DefSetExtractor] load %s (%s) (%s)", pkg.Name, pkg.PkgPath, pkg.ID)
		defs[i] = s.extractor.Extract(f)
	}
	logger.Debugf("[DefSetExtractor] %d files loaded", len(defs))
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
