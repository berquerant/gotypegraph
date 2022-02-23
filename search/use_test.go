package search_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/berquerant/gotypegraph/load"
	"github.com/berquerant/gotypegraph/search"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestPkg(t *testing.T) {
	t.Run("builtin", func(t *testing.T) {
		assert.True(t, search.NewBuiltinPkg().IsBuiltin())
	})
	t.Run("pkg", func(t *testing.T) {
		p := search.NewPkg(&packages.Package{
			Name:    "pkgname",
			PkgPath: "pkgpath",
		})
		assert.Equal(t, "pkgname", p.Name())
		assert.Equal(t, "pkgpath", p.Path())
	})
	t.Run("pkg with name", func(t *testing.T) {
		p := search.NewPkgWithName("pkgname", "pkgpath")
		assert.Equal(t, "pkgname", p.Name())
		assert.Equal(t, "pkgpath", p.Path())
	})
}

type useSearcherResult struct {
	refPkg      string
	refIdent    string
	refNodeType search.NodeType
	refASTIdent string
	refRecv     string
	defPkg      string
	defIdent    string
	defNodeType search.NodeType
	defRecv     string
}

func (s *useSearcherResult) test(t *testing.T, use search.Use) {
	var (
		r = use.Ref()
		d = use.Def()
	)
	assert.Equal(t, s.refPkg, r.Pkg().Name(), "ref pkg")
	assert.Equal(t, s.refIdent, r.Ident().Name, "ref ident")
	assert.Equal(t, s.refNodeType, r.Type(), "ref node type")
	assert.Equal(t, s.refASTIdent, r.Name(), "ref ast ident")
	assert.Equal(t, s.refRecv, r.RecvString(), "ref recv")
	assert.Equal(t, s.defPkg, d.Pkg().Name(), "def pkg")
	assert.Equal(t, s.defIdent, d.Name(), "def ident")
	assert.Equal(t, s.defNodeType, d.Type(), "def node type")
	assert.Equal(t, s.defRecv, d.RecvString(), "def recv")
}

func TestUseSearcher(t *testing.T) {
	pkgs, err := load.New().Load("./testpkg/...")
	assert.Nil(t, err)
	var (
		searcher = newUseSearcher(pkgs)
		got      = []search.Use{}
	)
	for r := range searcher.Search() {
		got = append(got, r)
	}
	sort.Slice(got, func(i, j int) bool {
		left, right := got[i].Ref(), got[j].Ref()
		return left.Ident().Pos() < right.Ident().Pos()
	})
	want := []*useSearcherResult{
		{
			refPkg:      "testpkg",
			refIdent:    "C1",
			refASTIdent: "V2",
			refNodeType: search.VarNodeType,
			defPkg:      "testpkg",
			defIdent:    "C1",
			defNodeType: search.ConstNodeType,
		},
		{
			refPkg:      "testpkg",
			refIdent:    "X",
			refASTIdent: "Y",
			refNodeType: search.TypeNodeType,
			defPkg:      "testpkg",
			defIdent:    "X",
			defNodeType: search.TypeNodeType,
		},
		{
			refPkg:      "testpkg",
			refIdent:    "V1",
			refASTIdent: "SameNameFunc",
			refNodeType: search.MethodNodeType,
			refRecv:     "*Y",
			defPkg:      "testpkg",
			defIdent:    "V1",
			defNodeType: search.VarNodeType,
		},
		{
			refPkg:      "testpkg",
			refIdent:    "V1",
			refASTIdent: "SameNameFunc",
			refNodeType: search.MethodNodeType,
			refRecv:     "*X",
			defPkg:      "testpkg",
			defIdent:    "V1",
			defNodeType: search.VarNodeType,
		},
		{
			refPkg:      "testpkg",
			refIdent:    "SameNameFunc",
			refASTIdent: "SameNameFunc",
			refNodeType: search.FuncNodeType,
			defPkg:      "sub",
			defIdent:    "SameNameFunc",
			defNodeType: search.FuncNodeType,
		},
	}
	assert.Equal(t, len(want), len(got))
	for i, r := range got {
		i := i
		r := r
		w := want[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			w.test(t, r)
		})
	}
}

func newUseSearcher(pkgs []*packages.Package) search.UseSearcher {
	defSetExtractor := search.NewDefSetExtractor(search.NewDefExtractor())
	defSetList := make([]search.DefSet, len(pkgs))
	for i, pkg := range pkgs {
		defSetList[i] = defSetExtractor.Extract(pkg)
	}
	return search.NewUseSearcher(
		pkgs,
		search.NewRefPkgSearcher(search.NewRefSearcher(), defSetList),
		search.NewObjExtractor(),
		search.NewTargetExtractor(),
		search.DefSetFilter(defSetList),
	)
}
