package search

import (
	"go/token"
	"go/types"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

type (
	FieldSearcher interface {
		Search(pkg *types.Package, pos token.Pos) (*types.TypeName, bool)
	}

	fieldSearcherRecord struct {
		typeName *types.TypeName
		start    token.Pos
		end      token.Pos
	}

	fieldSearcher struct {
		d map[string][]*fieldSearcherRecord
	}
)

func NewFieldSearcher(pkgs []*types.Package) FieldSearcher {
	d := map[string][]*fieldSearcherRecord{}
	for _, pkg := range pkgs {
		id := pkg.Path()
		if _, found := d[id]; found {
			continue
		}
		d[id] = append(d[id], extractFieldSearcherRecords(pkg)...)
	}
	return &fieldSearcher{
		d: d,
	}
}

func extractFieldSearcherRecords(pkg *types.Package) []*fieldSearcherRecord {
	var records []*fieldSearcherRecord
	for _, nm := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(nm)
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		nmd, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		if str, ok := nmd.Underlying().(*types.Struct); ok {
			if str.NumFields() == 0 {
				continue
			}
			// pos range of the struct fields
			var (
				start = str.Field(0).Pos()
				end   = str.Field(str.NumFields() - 1).Pos()
			)
			logger.Debugf("[FieldSearcher][init] %s (%s) %s from %d to %d",
				pkg.Name(), pkg.Path(), typeName.Name(), start, end)
			records = append(records, &fieldSearcherRecord{
				typeName: typeName,
				start:    start,
				end:      end,
			})
		}
	}
	return records
}

func (s *fieldSearcherRecord) contains(p token.Pos) bool { return s.start <= p && p <= s.end }

func (s *fieldSearcher) Search(pkg *types.Package, pos token.Pos) (*types.TypeName, bool) {
	if pkg == nil {
		return nil, false
	}
	rs, ok := s.d[pkg.Path()]
	if !ok {
		return nil, false
	}
	for _, r := range rs {
		if r.contains(pos) {
			return r.typeName, true
		}
	}
	return nil, false
}

type (
	FieldSearcherBuilder interface {
		Add(*types.Package)
		Build() FieldSearcher
	}

	fieldSearcherBuilder struct {
		pkgs map[string]*types.Package
	}
)

func NewFieldSearcherBuilder() FieldSearcherBuilder {
	return &fieldSearcherBuilder{
		pkgs: map[string]*types.Package{},
	}
}

func (s *fieldSearcherBuilder) Add(pkg *types.Package) { s.pkgs[pkg.Path()] = pkg }
func (s *fieldSearcherBuilder) Build() FieldSearcher {
	var (
		pkgs = make([]*types.Package, len(s.pkgs))
		i    int
	)
	for _, pkg := range s.pkgs {
		pkgs[i] = pkg
		i++
	}
	return NewFieldSearcher(pkgs)
}

func NewFieldSearcherFromPackages(pkgs []*packages.Package) FieldSearcher {
	builder := NewFieldSearcherBuilder()
	for _, pkg := range pkgs {
		for _, obj := range pkg.TypesInfo.Defs {
			if obj != nil {
				if p := obj.Pkg(); p != nil {
					builder.Add(p)
				}
			}
		}
		for _, obj := range pkg.TypesInfo.Uses {
			if obj != nil {
				if p := obj.Pkg(); p != nil {
					builder.Add(p)
				}
			}
		}
	}
	return builder.Build()
}
