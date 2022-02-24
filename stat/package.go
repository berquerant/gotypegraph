package stat

import (
	"encoding/json"
	"fmt"

	"github.com/berquerant/gotypegraph/search"
)

type (
	Pkg interface {
		ID() string
		Pkg() search.Pkg
	}

	pkg struct {
		pkg search.Pkg
	}
)

func NewPkg(p search.Pkg) Pkg {
	return &pkg{
		pkg: p,
	}
}
func (s *pkg) Pkg() search.Pkg { return s.pkg }
func (s *pkg) ID() string      { return s.pkg.Path() }

/* package stat of dependencies */

type (
	PkgStatSet interface {
		Stats() []PkgStat
		Get(Pkg) (PkgStat, bool)
	}
	PkgStat interface {
		Pkg() Pkg
		Weight() int
		Refs() PkgStatCell
		Defs() PkgStatCell
	}
	PkgStatCell interface {
		Pkg() Pkg
		Deps() []PkgStatDep
		Get(Pkg) (PkgStatDep, bool)
	}
	PkgStatDep interface {
		Pkg() Pkg
		Weight() int
	}
	PkgStatCalculator interface {
		Add(ref, def Pkg)
		Result() PkgStatSet
	}
)

type pkgStatSet struct {
	stats map[string]*pkgStat
}

func (s *pkgStatSet) MarshalJSON() ([]byte, error) { return json.Marshal(s.stats) }

func (s *pkgStatSet) Stats() []PkgStat {
	var (
		i     int
		stats = make([]PkgStat, len(s.stats))
	)
	for _, x := range s.stats {
		stats[i] = x
		i++
	}
	return stats
}

func (s *pkgStatSet) Get(pkg Pkg) (PkgStat, bool) {
	x, ok := s.stats[pkg.ID()]
	return x, ok
}

func newPkgStat(pkg Pkg) *pkgStat {
	return &pkgStat{
		pkg:  pkg,
		defs: newPkgStatCell(pkg),
		refs: newPkgStatCell(pkg),
	}
}

type pkgStat struct {
	pkg  Pkg
	defs *pkgStatCell
	refs *pkgStatCell
}

func (s *pkgStat) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"pkg":    s.pkg.ID(),
		"defs":   s.defs,
		"refs":   s.refs,
		"weight": s.Weight(),
	})
}
func (s *pkgStat) Pkg() Pkg          { return s.pkg }
func (s *pkgStat) Defs() PkgStatCell { return s.defs }
func (s *pkgStat) Refs() PkgStatCell { return s.refs }
func (s *pkgStat) Weight() int {
	var n int
	for _, x := range s.defs.Deps() {
		n += x.Weight()
	}
	for _, x := range s.refs.Deps() {
		n += x.Weight()
	}
	return n
}

func newPkgStatDep(pkg Pkg) *pkgStatDep {
	return &pkgStatDep{
		pkg:    pkg,
		weight: 1,
	}
}

type pkgStatDep struct {
	pkg    Pkg
	weight int
}

func (s *pkgStatDep) Pkg() Pkg    { return s.pkg }
func (s *pkgStatDep) Weight() int { return s.weight }
func (s *pkgStatDep) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"pkg":    s.pkg.ID(),
		"weight": s.weight,
	})
}

func newPkgStatCell(pkg Pkg) *pkgStatCell {
	return &pkgStatCell{
		pkg:  pkg,
		deps: map[string]*pkgStatDep{},
	}
}

type pkgStatCell struct {
	pkg  Pkg
	deps map[string]*pkgStatDep
}

func (s *pkgStatCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"pkg":  s.pkg.ID(),
		"deps": s.deps,
	})
}
func (s *pkgStatCell) add(pkg Pkg) {
	if x, ok := s.deps[pkg.ID()]; ok {
		x.weight++
		return
	}
	s.deps[pkg.ID()] = newPkgStatDep(pkg)
}
func (s *pkgStatCell) Pkg() Pkg { return s.pkg }
func (s *pkgStatCell) Deps() []PkgStatDep {
	var (
		deps = make([]PkgStatDep, len(s.deps))
		i    int
	)
	for _, dep := range s.deps {
		deps[i] = dep
		i++
	}
	return deps
}
func (s *pkgStatCell) Get(pkg Pkg) (PkgStatDep, bool) {
	dep, ok := s.deps[pkg.ID()]
	return dep, ok
}

func NewPkgStatCalculator() PkgStatCalculator {
	return &pkgStatCalculator{
		d: map[string]*pkgStat{},
	}
}

type pkgStatCalculator struct {
	d map[string]*pkgStat
}

func (s *pkgStatCalculator) Add(ref, def Pkg) {
	if x, ok := s.d[ref.ID()]; ok {
		x.refs.add(def)
	} else {
		x := newPkgStat(ref)
		x.refs.add(def)
		s.d[ref.ID()] = x
	}
	if x, ok := s.d[def.ID()]; ok {
		x.defs.add(ref)
	} else {
		x := newPkgStat(def)
		x.defs.add(ref)
		s.d[def.ID()] = x
	}
}

func (s *pkgStatCalculator) Result() PkgStatSet {
	return &pkgStatSet{
		stats: s.d,
	}
}

/* package dependencies */

type (
	PkgDep interface {
		Ref() Pkg
		Def() Pkg
		Weight() int
	}

	pkgDep struct {
		ref    Pkg
		def    Pkg
		weight int
	}

	PkgDepCalculator interface {
		Add(ref, def Pkg)
		Result() []PkgDep
	}
)

func (s *pkgDep) Ref() Pkg    { return s.ref }
func (s *pkgDep) Def() Pkg    { return s.def }
func (s *pkgDep) Weight() int { return s.weight }
func (s *pkgDep) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"ref":    s.ref.ID(),
		"def":    s.def.ID(),
		"weight": s.weight,
	})
}

func NewPkgDepCalculator() PkgDepCalculator {
	return &pkgDepCalculator{
		d: make(map[string]*pkgDep),
	}
}

type pkgDepCalculator struct {
	d map[string]*pkgDep
}

func (*pkgDepCalculator) id(ref, dep Pkg) string {
	return fmt.Sprintf("%s>%s", ref.ID(), dep.ID())
}

func (s *pkgDepCalculator) Add(ref, def Pkg) {
	id := s.id(ref, def)
	if dep, found := s.d[id]; found {
		dep.weight++
		return
	}
	s.d[id] = &pkgDep{
		ref:    ref,
		def:    def,
		weight: 1,
	}
}

func (s *pkgDepCalculator) Result() []PkgDep {
	var (
		i int
		r = make([]PkgDep, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}
