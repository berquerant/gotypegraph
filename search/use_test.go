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

type useSearcherTestcase struct {
	title string
	opt   []search.UseSearcherOption
	want  []*useSearcherResult
}

func (s *useSearcherTestcase) test(t *testing.T, pkgs []*packages.Package) {
	got := doUseSearch(pkgs, s.opt...)
	assert.Equal(t, len(s.want), len(got))
	for i, r := range got {
		i := i
		r := r
		w := s.want[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			w.test(t, r)
		})
	}
}

func TestUseSearcher(t *testing.T) {
	pkgs, err := load.New().Load("./testpkg/...")
	assert.Nil(t, err)

	for _, tc := range []useSearcherTestcase{
		{
			title: "with foreign ignore same pkg",
			opt: []search.UseSearcherOption{
				search.WithUseSearcherSearchForeign(true),
				search.WithUseSearcherIgnorePkgSelfloop(true),
			},
			want: []*useSearcherResult{
				{
					refPkg:      "sub",
					refIdent:    "Fprintln",
					refASTIdent: "SameNameFunc",
					refNodeType: search.FuncNodeType,
					defPkg:      "fmt",
					defIdent:    "Fprintln",
					defNodeType: search.FuncNodeType,
				},
				{
					refPkg:      "sub",
					refIdent:    "Stderr",
					refASTIdent: "SameNameFunc",
					refNodeType: search.FuncNodeType,
					defPkg:      "os",
					defIdent:    "Stderr",
					defNodeType: search.VarNodeType,
				},
				{
					refPkg:      "testpkg",
					refIdent:    "Println",
					refASTIdent: "SameNameFunc",
					refNodeType: search.MethodNodeType,
					refRecv:     "*X",
					defPkg:      "fmt",
					defIdent:    "Println",
					defNodeType: search.FuncNodeType,
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
			},
		},
		{
			title: "with foreign",
			opt: []search.UseSearcherOption{
				search.WithUseSearcherSearchForeign(true),
			},
			want: []*useSearcherResult{
				{
					refPkg:      "sub",
					refIdent:    "Fprintln",
					refASTIdent: "SameNameFunc",
					refNodeType: search.FuncNodeType,
					defPkg:      "fmt",
					defIdent:    "Fprintln",
					defNodeType: search.FuncNodeType,
				},
				{
					refPkg:      "sub",
					refIdent:    "Stderr",
					refASTIdent: "SameNameFunc",
					refNodeType: search.FuncNodeType,
					defPkg:      "os",
					defIdent:    "Stderr",
					defNodeType: search.VarNodeType,
				},
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
					refIdent:    "Println",
					refASTIdent: "SameNameFunc",
					refNodeType: search.MethodNodeType,
					refRecv:     "*X",
					defPkg:      "fmt",
					defIdent:    "Println",
					defNodeType: search.FuncNodeType,
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
			},
		},
		{
			title: "no opt",
			want: []*useSearcherResult{
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
			},
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			tc.test(t, pkgs)
		})
	}
}

func doUseSearch(pkgs []*packages.Package, opt ...search.UseSearcherOption) []search.Use {
	defSetExtractor := search.NewDefSetExtractor(search.NewDefExtractor())
	defSetList := make([]search.DefSet, len(pkgs))
	for i, pkg := range pkgs {
		defSetList[i] = defSetExtractor.Extract(pkg)
	}
	var (
		searcher = search.NewUseSearcher(
			pkgs,
			search.NewRefPkgSearcher(search.NewRefSearcher(), defSetList),
			search.NewObjExtractor(),
			search.NewTargetExtractor(),
			search.NewFieldSearcherFromPackages(pkgs),
			search.DefSetFilter(defSetList),
			opt...,
		)
		got = []search.Use{}
	)
	for r := range searcher.Search() {
		got = append(got, r)
	}
	sort.Slice(got, func(i, j int) bool {
		left, right := got[i].Ref(), got[j].Ref()
		return left.Ident().Pos() < right.Ident().Pos()
	})
	return got
}
