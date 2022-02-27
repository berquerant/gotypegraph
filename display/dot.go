package display

import (
	"fmt"
	"sort"

	"github.com/berquerant/gotypegraph/util"
)

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
	r util.Ranking
	p util.Percentiler
}

func (s *attrRanking) get(v int) int { return s.p.Percentile(s.r.GetPercentile(v)) }

func newFontsizeRanking(r util.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: util.NewPercentiler(minFontsize, maxFontsize),
	}
}

func newPenwidthRanking(r util.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: util.NewPercentiler(minPenwidth, maxPenwidth),
	}
}

func newWeightRanking(r util.Ranking) *attrRanking {
	return &attrRanking{
		r: r,
		p: util.NewPercentiler(minWeight, maxWeight),
	}
}

func generateNodeLabelHTML(
	titleKey, titleValue string,
	ref, def, uniqRef, uniqDef int,
) string {
	return fmt.Sprintf(`<table border="0">
  <tr>
    <td><b>%s</b></td>
    <td><b>%s</b></td>
  </tr>
  <tr>
    <td align="left"><b>RefDef</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>Ref</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>Def</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>UniqRef</b></td>
    <td align="right">%d</td>
  </tr>
  <tr>
    <td align="left"><b>UniqDef</b></td>
    <td align="right">%d</td>
  </tr>
</table>`, titleKey, titleValue, ref+def, ref, def, uniqRef, uniqDef)
}

type (
	refDefTooltipElem struct {
		id     string
		weight int
	}
	refDefTooltip struct {
		title string
		refs  []*refDefTooltipElem
		defs  []*refDefTooltipElem
	}
)

func newRefDefTooltipElem(id string, weight int) *refDefTooltipElem {
	return &refDefTooltipElem{
		id:     id,
		weight: weight,
	}
}

func (s *refDefTooltipElem) String() string { return fmt.Sprintf("%s %d", s.id, s.weight) }

func newRefDefTooltip(title string) *refDefTooltip {
	return &refDefTooltip{
		title: title,
		refs:  []*refDefTooltipElem{},
		defs:  []*refDefTooltipElem{},
	}
}

func (s *refDefTooltip) addRef(elem *refDefTooltipElem) { s.refs = append(s.refs, elem) }
func (s *refDefTooltip) addDef(elem *refDefTooltipElem) { s.defs = append(s.defs, elem) }
func (s *refDefTooltip) String() string {
	var b util.StringBuilder
	b.Writeln(s.title)
	sort.Slice(s.refs, func(i, j int) bool { return s.refs[i].id < s.refs[j].id })
	sort.Slice(s.defs, func(i, j int) bool { return s.defs[i].id < s.defs[j].id })
	b.Writeln("Ref:")
	for _, x := range s.refs {
		b.Writeln(x.String())
	}
	b.Writeln("Def:")
	for _, x := range s.defs {
		b.Writeln(x.String())
	}
	return b.String()
}
