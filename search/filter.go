package search

import (
	"go/token"

	"github.com/berquerant/gotypegraph/util"
	"golang.org/x/tools/go/packages"
)

// Filter selects a target.
type Filter func(Target) bool

func (s Filter) And(next Filter) Filter {
	return func(tgt Target) bool {
		return s(tgt) && next(tgt)
	}
}

func (s Filter) Or(next Filter) Filter {
	return func(tgt Target) bool {
		return s(tgt) || next(tgt)
	}
}

// UniverseFilter selects a builtin target.
func UniverseFilter(tgt Target) bool {
	return tgt.Obj() != nil && tgt.Obj().Pkg() == nil
}

// ExportedFilter selects an exported target.
func ExportedFilter(tgt Target) bool {
	return tgt.Obj() != nil && tgt.Obj().Exported()
}

// ObjectNameFilter selects a target whose object name matched.
func ObjectNameFilter(pair util.RegexpPair) Filter {
	return func(tgt Target) bool {
		return tgt.Obj() != nil && pair.MatchString(tgt.Obj().Name())
	}
}

// PkgNameFilter selects a target whose package name matched.
func PkgNameFilter(pair util.RegexpPair) Filter {
	return func(tgt Target) bool {
		return tgt.Obj() != nil && tgt.Obj().Pkg() != nil && pair.MatchString(tgt.Obj().Pkg().Name())
	}
}

// OtherPkgFilter selects a target whose package name is not matched with given packages.
func OtherPkgFilter(pkgs []*packages.Package) Filter {
	ss := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		ss[i] = pkg.PkgPath
	}
	pkgSet := util.NewStringSet(ss...)

	return func(tgt Target) bool {
		return tgt.Obj() != nil && tgt.Obj().Pkg() != nil && pkgSet.In(tgt.Obj().Pkg().Path())
	}
}

// DefSetFilter selects a target belongs to the defs.
func DefSetFilter(setList []DefSet) Filter {
	pkgSet := make(map[string]map[token.Pos]bool, len(setList))
	for _, defSet := range setList {
		posSet := make(map[token.Pos]bool)
		for _, def := range defSet.Defs() {
			for _, vs := range def.ValueSpecs() {
				for _, nm := range vs.Names {
					posSet[nm.Pos()] = true
				}
			}
			for _, fd := range def.FuncDecls() {
				posSet[fd.Pos()] = true
			}
			for _, ts := range def.TypeSpecs() {
				posSet[ts.Pos()] = true
			}
		}
		pkgSet[defSet.Pkg().PkgPath] = posSet
	}

	return func(tgt Target) bool {
		obj := tgt.Obj()
		if obj == nil || obj.Pkg() == nil {
			return false
		}
		p, ok := pkgSet[obj.Pkg().Path()]
		return ok && p[obj.Pos()]
	}
}
