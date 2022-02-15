package display

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/berquerant/gotypegraph/stat"
	"github.com/berquerant/gotypegraph/use"
	"github.com/berquerant/gotypegraph/util"
)

type statDotWriter struct {
	w io.Writer
	g stat.Graph
}

func NewStatDotWriter(w io.Writer) Writer {
	return &statDotWriter{
		w: w,
		g: stat.NewGraph(),
	}
}

func (s *statDotWriter) Write(result *use.Result) error {
	s.g.Add(stat.NewPair(result))
	return nil
}

func (*statDotWriter) newTooltip(pkg *stat.PkgWeight, deps []*stat.PkgDep) string {
	var b util.StringBuilder
	b.Writeln(pkg.Pkg)
	b.Writeln(pkg.Position)
	var (
		inDeps  = []*stat.PkgDep{}
		outDeps = []*stat.PkgDep{}
	)
	for _, dep := range deps {
		if dep.Def == pkg.Pkg {
			inDeps = append(inDeps, dep)
		}
		if dep.Ref == pkg.Pkg {
			outDeps = append(outDeps, dep)
		}
	}
	sort.Slice(inDeps, func(i, j int) bool { return inDeps[i].Ref < inDeps[j].Ref })
	sort.Slice(outDeps, func(i, j int) bool { return outDeps[i].Def < outDeps[j].Def })

	b.Writeln("In:")
	for _, dep := range inDeps {
		b.Writelnf("%s %d", dep.Ref, dep.Weight)
	}
	b.Writeln("Out:")
	for _, dep := range outDeps {
		b.Writelnf("%s %d", dep.Def, dep.Weight)
	}
	return b.String()
}

func (s *statDotWriter) build() string {
	var (
		g               = dot.NewGraph("G")
		pkgIOs          = s.g.PkgIOs()
		pkgs            = s.g.Pkgs()
		fontsizeRanking = s.newFontsizeRanking(pkgs)
		deps            = s.g.PkgDeps()
	)
	for _, pkg := range pkgs {
		var (
			n     = dot.NewNode(pkg.Pkg)
			ios   = pkgIOs[pkg.Pkg]
			label = fmt.Sprintf("<\n%s\n>", generateNodeLabelHTML(
				"package", pkg.Pkg,
				ios.In, ios.Out, ios.UniqIn(), ios.UniqOut()))
			tooltip  = s.newTooltip(pkg, deps)
			fontsize = strconv.Itoa(fontsizeRanking.get(pkg.Weight))
		)
		n.Attrs.Add("shape", "box").
			Add("label", label, dot.WithAsRaw(true)).
			Add("tooltip", tooltip).
			Add("fontsize", fontsize)
		g.Nodes.Add(n)
	}

	var (
		penwidthRanking = s.newPenwidthRanking(deps)
		weightRanking   = s.newWeightRanking(deps)
	)
	for _, dep := range deps {
		var (
			e        = dot.NewEdge(dep.Ref, dep.Def)
			tooltip  = fmt.Sprintf("%s -> %s [%d]", dep.Ref, dep.Def, dep.Weight)
			penwidth = strconv.Itoa(penwidthRanking.get(dep.Weight))
			weight   = strconv.Itoa(weightRanking.get(dep.Weight))
		)
		e.Attrs.Add("label", strconv.Itoa(dep.Weight)).
			Add("tooltip", tooltip).
			Add("labeltooltip", tooltip).
			Add("penwidth", penwidth).
			Add("weight", weight)
		g.Edges.Add(e)
	}
	return g.String()
}

func (s *statDotWriter) Flush() error {
	if _, err := fmt.Fprintln(s.w, s.build()); err != nil {
		return fmt.Errorf("StatDotWriter: %w", err)
	}
	return nil
}

func (*statDotWriter) newFontsizeRanking(pkgs []*stat.PkgWeight) *attrRanking {
	r := stat.NewRanking()
	for _, pkg := range pkgs {
		r.Add(pkg.Weight)
	}
	return newFontsizeRanking(r)
}

func (*statDotWriter) newPenwidthRanking(deps []*stat.PkgDep) *attrRanking {
	r := stat.NewRanking()
	for _, dep := range deps {
		r.Add(dep.Weight)
	}
	return newPenwidthRanking(r)
}

func (*statDotWriter) newWeightRanking(deps []*stat.PkgDep) *attrRanking {
	r := stat.NewRanking()
	for _, dep := range deps {
		r.Add(dep.Weight)
	}
	return newWeightRanking(r)
}

type dotWriter struct {
	w io.Writer
	g stat.Graph
}

func NewDotWriter(w io.Writer) Writer {
	return &dotWriter{
		w: w,
		g: stat.NewGraph(),
	}
}

func (s *dotWriter) Write(result *use.Result) error {
	s.g.Add(stat.NewPair(result))
	return nil
}

const (
	subgraphFontsize = 24
)

func (*dotWriter) newTooltip(node *stat.NodeWeight, deps []*stat.Dep) string {
	var b util.StringBuilder
	if node.Node.Recv != "" {
		b.WriteString(fmt.Sprintf("(%s) ", node.Node.Recv))
	}
	b.Writelnf("%s.%s", node.Node.Pkg, node.Node.Name)
	b.Writeln(node.Node.Position.String())
	var (
		inDeps  = []*stat.Dep{}
		outDeps = []*stat.Dep{}
		toID    = func(n *stat.Node) string { return fmt.Sprintf("%s.%s", n.Pkg, n.Name) }
		nid     = toID(node.Node)
	)
	for _, dep := range deps {
		if toID(dep.Pair.Def) == nid {
			inDeps = append(inDeps, dep)
		}
		if toID(dep.Pair.Ref) == nid {
			outDeps = append(outDeps, dep)
		}
	}
	sort.Slice(inDeps, func(i, j int) bool { return toID(inDeps[i].Pair.Ref) < toID(inDeps[j].Pair.Ref) })
	sort.Slice(outDeps, func(i, j int) bool { return toID(outDeps[i].Pair.Def) < toID(outDeps[j].Pair.Def) })

	b.Writeln("In:")
	for _, dep := range inDeps {
		b.Writelnf("%s %d", toID(dep.Pair.Ref), dep.Weight)
	}
	b.Writeln("Out:")
	for _, dep := range outDeps {
		b.Writelnf("%s %d", toID(dep.Pair.Def), dep.Weight)
	}
	return b.String()
}

func (s *dotWriter) build() string {
	var (
		g               = dot.NewGraph("G")
		pkgMap          = s.g.PkgToNodes()
		fontsizeRanking = s.newFontsizeRanking(pkgMap)
		deps            = s.g.Deps()
	)

	g.Attrs.Add("newrank", "true") // due to triangulation failed on generating layout
	for pkg, nodes := range pkgMap {
		sg := dot.NewSubgraph(pkg, true)
		sg.Attrs.Add("color", "lightgrey").
			Add("style", "filled").
			Add("label", pkg).
			Add("fontsize", strconv.Itoa(subgraphFontsize))
		nodeIOs := s.g.NodeIOs()
		for _, node := range nodes {
			var (
				n      = dot.NewNode(s.nodeToID(node.Node))
				ios, _ = nodeIOs.Get(node.Node)
				label  = fmt.Sprintf("<\n%s\n>", generateNodeLabelHTML(
					node.Node.Type.String(), s.nodeToTitleValue(node.Node),
					ios.In, ios.Out,
					ios.UniqIn(), ios.UniqOut()))
				tooltip  = s.newTooltip(node, deps)
				fontsize = strconv.Itoa(fontsizeRanking.get(node.Weight))
			)
			n.Attrs.Add("color", "white").
				Add("style", "filled").
				Add("shape", "box").
				Add("label", label, dot.WithAsRaw(true)).
				Add("tooltip", tooltip).
				Add("fontsize", fontsize)
			sg.Nodes.Add(n)
		}
		g.Subgraphs.Add(sg)
	}

	var (
		penwidthRanking = s.newPenwidthRanking(deps)
		weightRanking   = s.newWeightRanking(deps)
	)

	for _, dep := range deps {
		e := dot.NewEdge(s.nodeToID(dep.Pair.Ref), s.nodeToID(dep.Pair.Def))
		if dep.Weight > 1 {
			e.Attrs.Add("label", strconv.Itoa(dep.Weight))
		}
		var (
			tooltip = fmt.Sprintf("%s.%s -> %s.%s [%d]",
				dep.Pair.Ref.Pkg, dep.Pair.Ref.Name,
				dep.Pair.Def.Pkg, dep.Pair.Def.Name,
				dep.Weight)
			penwidth = strconv.Itoa(penwidthRanking.get(dep.Weight))
			weight   = strconv.Itoa(weightRanking.get(dep.Weight))
		)
		e.Attrs.Add("tooltip", tooltip).
			Add("labeltooltip", tooltip).
			Add("penwidth", penwidth).
			Add("weight", weight)
		g.Edges.Add(e)
	}
	return g.String()
}

func (s *dotWriter) Flush() error {
	if _, err := fmt.Fprintln(s.w, s.build()); err != nil {
		return fmt.Errorf("DotWriter: %w", err)
	}
	return nil
}

func (*dotWriter) nodeToID(node *stat.Node) string { return fmt.Sprintf("%s__%s", node.Pkg, node.Name) }
func (*dotWriter) nodeToTitleValue(node *stat.Node) string {
	if node.Recv == "" {
		return node.Name
	}
	return fmt.Sprintf("(%s).%s", node.Recv, node.Name)
}

func (*dotWriter) newFontsizeRanking(nodeMap map[string][]*stat.NodeWeight) *attrRanking {
	r := stat.NewRanking()
	for _, nodes := range nodeMap {
		for _, n := range nodes {
			r.Add(n.Weight)
		}
	}
	return newFontsizeRanking(r)
}

func (*dotWriter) newPenwidthRanking(deps []*stat.Dep) *attrRanking {
	r := stat.NewRanking()
	for _, dep := range deps {
		r.Add(dep.Weight)
	}
	return newPenwidthRanking(r)
}

func (*dotWriter) newWeightRanking(deps []*stat.Dep) *attrRanking {
	r := stat.NewRanking()
	for _, dep := range deps {
		r.Add(dep.Weight)
	}
	return newWeightRanking(r)
}

/* rankings */

const (
	minFontsize = 8
	maxFontsize = 24
	minPenwidth = 1
	maxPenwidth = 5
	minWeight   = 1
	maxWeight   = 100
)

type attrRanking struct {
	r stat.Ranking
	p stat.Percentiler
}

func (s *attrRanking) get(v int) int { return s.p.Get(s.r.GetPercentile(v)) }

func newFontsizeRanking(r stat.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: stat.NewPercentiler(minFontsize, maxFontsize),
	}
}

func newPenwidthRanking(r stat.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: stat.NewPercentiler(minPenwidth, maxPenwidth),
	}
}

func newWeightRanking(r stat.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: stat.NewPercentiler(minWeight, maxWeight),
	}
}

/* label generator */

func generateNodeLabelHTML(titleKey, titleValue string, in, out, uniqIn, uniqOut int) string {
	return fmt.Sprintf(`<table border="0">
  <tr>
    <td><b>%s</b></td>
    <td><b>%s</b></td>
  </tr>
  <tr>
    <td align="left"><b>IO</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>In</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>Out</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>UniqIn</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>UniqOut</b></td>
    <td align="right">%d</td>
  </tr>
</table>`, titleKey, titleValue, in+out, in, out, uniqIn, uniqOut)
}
