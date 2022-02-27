package display

import (
	"fmt"
	"io"
	"strconv"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/berquerant/gotypegraph/search"
	"github.com/berquerant/gotypegraph/stat"
	"github.com/berquerant/gotypegraph/util"
)

type packageDotWriter struct {
	w           io.Writer
	depCalc     stat.PkgDepCalculator
	statDepCalc stat.PkgStatCalculator
}

func NewPackageDotWriter(w io.Writer) Writer {
	return &packageDotWriter{
		w:           w,
		depCalc:     stat.NewPkgDepCalculator(),
		statDepCalc: stat.NewPkgStatCalculator(),
	}
}

func (s *packageDotWriter) Write(node search.Use) error {
	var (
		ref = stat.NewPkg(node.Ref().Pkg())
		def = stat.NewPkg(node.Def().Pkg())
	)
	s.depCalc.Add(ref, def)
	s.statDepCalc.Add(ref, def)
	return nil
}

func (s *packageDotWriter) Flush() error {
	if _, err := fmt.Fprintln(s.w, s.build().String()); err != nil {
		return fmt.Errorf("PackageDotWriter: %w", err)
	}
	return nil
}

func (s *packageDotWriter) build() dot.Graph {
	var (
		deps  = s.depCalc.Result()
		stats = s.statDepCalc.Result()

		nodeList = dot.NewNodeList()

		fontsizeRanking = s.fontsizeRanking(stats)
	)

	for _, stat := range stats.Stats() {
		var (
			fontsize = fontsizeRanking.get(stat.Weight())
			label    = s.nodeLabel(stat)
			tooltip  = s.nodeTooltip(stat)
			attrList = dot.NewAttrList().
					Add(dot.NewAttr("shape", "box")).
					Add(dot.NewAttr("label", label, dot.WithAttrRaw(true))).
					Add(dot.NewAttr("tooltip", tooltip)).
					Add(dot.NewAttr("fontsize", strconv.Itoa(fontsize)))
			node = dot.NewNode(dot.ID(stat.Pkg().ID()), dot.WithNodeAttrList(attrList))
		)
		nodeList.Add(node)
	}

	var (
		edgeList = dot.NewEdgeList()

		penwidthRanking = s.penwidthRanking(deps)
		weightRanking   = s.weightRanking(deps)
	)

	for _, dep := range deps {
		var (
			penwidth = penwidthRanking.get(dep.Weight())
			weight   = weightRanking.get(dep.Weight())
			tooltip  = s.edgeTooltip(dep)
			label    = s.edgeLabel(dep)
			attrList = dot.NewAttrList().
					Add(dot.NewAttr("label", label)).
					Add(dot.NewAttr("tooltip", tooltip)).
					Add(dot.NewAttr("labeltooltip", tooltip)).
					Add(dot.NewAttr("penwidth", strconv.Itoa(penwidth))).
					Add(dot.NewAttr("weight", strconv.Itoa(weight)))
			edge = dot.NewEdge(
				dot.ID(dep.Ref().ID()),
				dot.ID(dep.Def().ID()),
				dot.WithEdgeAttrList(attrList),
			)
		)
		edgeList.Add(edge)
	}
	return dot.NewGraph("G", nodeList, edgeList)
}

func (*packageDotWriter) fontsizeRanking(stats stat.PkgStatSet) *attrRanking {
	r := util.NewRanking()
	for _, x := range stats.Stats() {
		r.Add(x.Weight())
	}
	return newFontsizeRanking(r)
}

func (*packageDotWriter) weightRanking(deps []stat.PkgDep) *attrRanking {
	r := util.NewRanking()
	for _, x := range deps {
		r.Add(x.Weight())
	}
	return newWeightRanking(r)
}

func (*packageDotWriter) penwidthRanking(deps []stat.PkgDep) *attrRanking {
	r := util.NewRanking()
	for _, x := range deps {
		r.Add(x.Weight())
	}
	return newPenwidthRanking(r)
}

func (*packageDotWriter) edgeLabel(dep stat.PkgDep) string {
	return strconv.Itoa(dep.Weight())
}

func (*packageDotWriter) edgeTooltip(dep stat.PkgDep) string {
	return fmt.Sprintf("%s -> %s [%d]", dep.Ref().Pkg().Path(), dep.Def().Pkg().Path(), dep.Weight())
}

func (*packageDotWriter) nodeLabel(pkgStat stat.PkgStat) string {
	return fmt.Sprintf("<\n%s\n>", generateNodeLabelHTML(
		"package", pkgStat.Pkg().Pkg().Name(),
		pkgStat.Refs().Weight(), pkgStat.Defs().Weight(),
		len(pkgStat.Refs().Deps()), len(pkgStat.Defs().Deps()),
	))
}

func (*packageDotWriter) nodeTooltip(pkgStat stat.PkgStat) string {
	tooltip := newRefDefTooltip(pkgStat.Pkg().Pkg().Path())
	for _, x := range pkgStat.Refs().Deps() {
		tooltip.addRef(newRefDefTooltipElem(x.Pkg().Pkg().Path(), x.Weight()))
	}
	for _, x := range pkgStat.Defs().Deps() {
		tooltip.addDef(newRefDefTooltipElem(x.Pkg().Pkg().Path(), x.Weight()))
	}
	return tooltip.String()
}
