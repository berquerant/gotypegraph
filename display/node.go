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

type nodeDotWriter struct {
	w           io.Writer
	depCalc     stat.NodeDepCalculator
	statDepCalc stat.NodeStatCalculator
	conf        *WriterConfig
}

func NewNodeDotWriter(w io.Writer, opt ...WriterOption) Writer {
	conf := newWriterConfig()
	for _, x := range opt {
		x(conf)
	}
	return &nodeDotWriter{
		w:           w,
		depCalc:     stat.NewNodeDepCalculator(),
		statDepCalc: stat.NewNodeStatCalculator(),
		conf:        conf,
	}
}

func (s *nodeDotWriter) Write(node search.Use) error {
	var (
		ref = stat.NewNode(node.Ref())
		def = stat.NewNode(node.Def())
	)
	s.statDepCalc.Add(ref, def)
	s.depCalc.Add(ref, def)
	return nil
}

func (s *nodeDotWriter) Flush() error {
	if _, err := fmt.Fprintln(s.w, s.build().String()); err != nil {
		return fmt.Errorf("NodeDotWriter: %w", err)
	}
	return nil
}

func (s *nodeDotWriter) build() dot.Graph {
	var (
		deps       = s.depCalc.Result()
		stats      = s.statDepCalc.Result()
		pkgStatMap = stat.NewNodeStatPkgMap(stats.Stats())

		subgraphList = dot.NewSubgraphList()

		fontsizeRanking    = s.fontsizeRanking(stats)
		pkgFontsizeRanking = s.pkgFontsizeRanking(pkgStatMap)
	)

	for _, pkg := range pkgStatMap.PkgList() {
		var (
			statList, _ = pkgStatMap.Get(pkg)
			nodeList    = dot.NewNodeList()
			pkgWeight   int
		)
		for _, st := range statList {
			pkgWeight += st.Weight()
			var (
				fontsize = fontsizeRanking.get(st.Weight())
				tooltip  = s.nodeTooltip(st)
				label    = s.nodeLabel(st)
				attrList = dot.NewAttrList().
						Add(dot.NewAttr("color", "white")).
						Add(dot.NewAttr("style", "filled")).
						Add(dot.NewAttr("shape", "box")).
						Add(dot.NewAttr("label", label, dot.WithAttrRaw(true))).
						Add(dot.NewAttr("tooltip", tooltip)).
						Add(dot.NewAttr("fontsize", strconv.Itoa(fontsize)))
				node = dot.NewNode(dot.ID(st.Node().ID()), dot.WithNodeAttrList(attrList))
			)
			nodeList.Add(node)
		}
		fontsize := pkgFontsizeRanking.get(pkgWeight)
		subgraph := dot.NewSubgraph(
			dot.ID(pkg.ID()),
			nodeList,
			dot.WithSubgraphCluster(true),
			dot.WithSubgraphAttrList(dot.NewAttrList().
				Add(dot.NewAttr("color", "lightgrey")).
				Add(dot.NewAttr("style", "filled")).
				Add(dot.NewAttr("label", pkg.Pkg().Name())).
				Add(dot.NewAttr("tooltip", pkg.Pkg().Path())).
				Add(dot.NewAttr("fontsize", strconv.Itoa(fontsize)))),
		)
		subgraphList.Add(subgraph)
	}

	var (
		edgeList = dot.NewEdgeList()

		penwidthRanking = s.penwidthRanking(deps)
		weightRanking   = s.weightRanking(deps)
	)

	for _, dep := range deps {
		var (
			penwidth  = penwidthRanking.get(dep.Weight())
			arrowsize = float64(penwidth) / 2
			weight    = weightRanking.get(dep.Weight())
			tooltip   = s.edgeTooltip(dep)
			attrList  = dot.NewAttrList().
					Add(dot.NewAttr("tooltip", tooltip)).
					Add(dot.NewAttr("labeltooltip", tooltip)).
					Add(dot.NewAttr("arrowsize", fmt.Sprint(arrowsize))).
					Add(dot.NewAttr("penwidth", strconv.Itoa(penwidth))).
					Add(dot.NewAttr("weight", strconv.Itoa(weight)))
		)
		if x, ok := s.edgeLabel(dep); ok {
			attrList.Add(dot.NewAttr("label", x))
		}
		edge := dot.NewEdge(
			dot.ID(dep.Ref().ID()),
			dot.ID(dep.Def().ID()),
			dot.WithEdgeAttrList(attrList),
		)
		edgeList.Add(edge)
	}
	return dot.NewGraph(
		"G",
		nil,
		edgeList,
		dot.WithGraphSubgraphList(subgraphList),
		dot.WithGraphAttrList(dot.NewAttrList().
			Add(dot.NewAttr("newrank", "true"))), // due to triangulation failed on generating layout
	)
}

func (*nodeDotWriter) edgeLabel(dep stat.NodeDep) (string, bool) {
	if dep.Weight() > 1 {
		return strconv.Itoa(dep.Weight()), true
	}
	return "", false
}

func (s *nodeDotWriter) edgeTooltip(dep stat.NodeDep) string {
	return fmt.Sprintf("%s.%s -> %s.%s [%d]",
		dep.Ref().Pkg().Pkg().Path(), s.nodeNameWithRecv(dep.Ref().Node()),
		dep.Def().Pkg().Pkg().Path(), s.nodeNameWithRecv(dep.Def().Node()),
		dep.Weight(),
	)
}

func (s *nodeDotWriter) nodeLabel(st stat.NodeStat) string {
	return fmt.Sprintf("<\n%s\n>",
		generateNodeLabelHTML(
			st.Node().Node().Type().String(), s.nodeToLabelTitle(st.Node().Node()),
			st.Refs().Weight(), st.Defs().Weight(),
			len(st.Refs().Deps()), len(st.Defs().Deps()),
		),
	)
}

func (s *nodeDotWriter) nodeToLabelTitle(node search.Node) string { return s.nodeNameWithRecv(node) }

func (*nodeDotWriter) nodeNameWithRecv(node search.Node) string {
	if recv := node.RecvString(); recv != "" {
		return fmt.Sprintf("(%s).%s", recv, node.Name())
	}
	return node.Name()
}

func (s *nodeDotWriter) nodeToTooltipID(node search.Node) string {
	return fmt.Sprintf("%s.%s", node.Pkg().Path(), s.nodeNameWithRecv(node))
}

func (s *nodeDotWriter) nodeToTooltipDetails(node search.Node) string {
	if pkg := node.Pkg().Pkg(); pkg != nil {
		return fmt.Sprintf("%s %s", s.nodeNameWithRecv(node), pkg.Fset.Position(node.Obj().Pos()))
	}
	return s.nodeToTooltipID(node)
}

func (s *nodeDotWriter) nodeTooltip(st stat.NodeStat) string { // TODO: char limit
	tooltip := newRefDefTooltip(s.nodeToTooltipDetails(st.Node().Node()))
	for _, x := range st.Refs().Deps() {
		tooltip.addRef(newRefDefTooltipElem(s.nodeToTooltipID(x.Node().Node()), x.Weight()))
	}
	for _, x := range st.Defs().Deps() {
		tooltip.addDef(newRefDefTooltipElem(s.nodeToTooltipID(x.Node().Node()), x.Weight()))
	}
	return tooltip.String()
}

func (s *nodeDotWriter) pkgFontsizeRanking(stats stat.NodeStatPkgMap) *attrRanking {
	r := util.NewRanking()
	for _, pkg := range stats.PkgList() {
		var w int
		statList, _ := stats.Get(pkg)
		for _, st := range statList {
			w += st.Weight()
		}
		r.Add(w)
	}
	return s.conf.newFontsizeRanking(r)
}

func (s *nodeDotWriter) fontsizeRanking(stats stat.NodeStatSet) *attrRanking {
	r := util.NewRanking()
	for _, x := range stats.Stats() {
		r.Add(x.Weight())
	}
	return s.conf.newFontsizeRanking(r)
}

func (s *nodeDotWriter) weightRanking(deps []stat.NodeDep) *attrRanking {
	r := util.NewRanking()
	for _, x := range deps {
		r.Add(x.Weight())
	}
	return s.conf.newWeightRanking(r)
}

func (s *nodeDotWriter) penwidthRanking(deps []stat.NodeDep) *attrRanking {
	r := util.NewRanking()
	for _, x := range deps {
		r.Add(x.Weight())
	}
	return s.conf.newPenwidthRanking(r)
}
