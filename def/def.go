package def

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

// Def is a set of the decls and the specs on the top level of the syntax.
type Def struct {
	ValueSpecs []*ast.ValueSpec // var, const
	FuncDecls  []*ast.FuncDecl  // func, method
	TypeSpecs  []*ast.TypeSpec  // type alias, defined type
}

type Extractor interface {
	// Extract extracts the top level decls and specs.
	Extract(f *ast.File) *Def
}

func NewExtactor() Extractor { return &extractor{} }

type extractor struct{}

func (*extractor) Extract(f *ast.File) *Def {
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
	return &Def{
		ValueSpecs: valueSpecs,
		FuncDecls:  funcDecls,
		TypeSpecs:  typeSpecs,
	}
}

// Set is a set of the definitions of the package.
type Set struct {
	Pkg  *packages.Package
	Defs []*Def
}

func NewSetExtractor(extractor Extractor) SetExtractor {
	return &setExtractor{
		extractor: extractor,
	}
}

type SetExtractor interface {
	Extract(p *packages.Package) *Set
}

type setExtractor struct {
	extractor Extractor
}

func (s *setExtractor) Extract(pkg *packages.Package) *Set {
	defs := make([]*Def, len(pkg.Syntax))
	for i, x := range pkg.Syntax {
		defs[i] = s.extractor.Extract(x)
	}
	return &Set{
		Pkg:  pkg,
		Defs: defs,
	}
}
