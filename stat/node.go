package stat

import (
	"encoding/json"
	"fmt"

	"github.com/berquerant/gotypegraph/search"
)

type (
	Node interface {
		ID() string
		Pkg() Pkg
		Node() search.Node
	}

	node struct {
		node search.Node
	}
)

func NewNode(n search.Node) Node {
	return &node{
		node: n,
	}
}

func (s *node) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{
		"pkg":  s.Pkg().ID(),
		"name": s.node.Name(),
		"type": s.node.Type().String(),
	}
	if recv := s.node.RecvString(); recv != "" {
		d["recv"] = recv
	}
	return json.Marshal(d)
}
func (s *node) Pkg() Pkg          { return NewPkg(s.node.Pkg()) }
func (s *node) Node() search.Node { return s.node }
func (s *node) ID() string {
	var (
		pkg  = s.Pkg().ID()
		nm   = s.node.Name()
		typ  = s.node.Type()
		recv = s.node.RecvString(search.WithNodeRawRecv(true))
	)
	if recv == "" {
		return fmt.Sprintf("%s-%d-%s", pkg, typ, nm)
	}
	return fmt.Sprintf("%s-%d-(%s)-%s", pkg, typ, recv, nm)
}

/* node stat of dependencies */

type (
	NodeStatPkgMap interface {
		PkgList() []Pkg
		Get(Pkg) ([]NodeStat, bool)
	}
	NodeStatSet interface {
		Stats() []NodeStat
		Get(Node) (NodeStat, bool)
	}
	NodeStat interface {
		Node() Node
		Weight() int
		Refs() NodeStatCell
		Defs() NodeStatCell
	}
	NodeStatCell interface {
		Node() Node
		Deps() []NodeStatDep
		Get(Node) (NodeStatDep, bool)
		Weight() int
	}
	NodeStatDep interface {
		Node() Node
		Weight() int
	}
	NodeStatCalculator interface {
		Add(ref, def Node)
		Result() NodeStatSet
	}
)

func NewNodeStatPkgMap(stats []NodeStat) NodeStatPkgMap {
	d := map[string][]NodeStat{}
	for _, x := range stats {
		id := x.Node().Pkg().ID()
		d[id] = append(d[id], x)
	}
	return &nodeStatPkgMap{
		d: d,
	}
}

type nodeStatPkgMap struct {
	d map[string][]NodeStat
}

func (s *nodeStatPkgMap) Get(pkg Pkg) ([]NodeStat, bool) {
	xs, ok := s.d[pkg.ID()]
	return xs, ok
}

func (s *nodeStatPkgMap) PkgList() []Pkg {
	pkgs := []Pkg{}
	for _, x := range s.d {
		if len(x) > 0 {
			pkgs = append(pkgs, x[0].Node().Pkg())
		}
	}
	return pkgs
}

type nodeStatSet struct {
	stats map[string]*nodeStat
}

func (s *nodeStatSet) MarshalJSON() ([]byte, error) { return json.Marshal(s.stats) }
func (s *nodeStatSet) Stats() []NodeStat {
	var (
		i     int
		stats = make([]NodeStat, len(s.stats))
	)
	for _, x := range s.stats {
		stats[i] = x
		i++
	}
	return stats
}
func (s *nodeStatSet) Get(node Node) (NodeStat, bool) {
	x, ok := s.stats[node.ID()]
	return x, ok
}

func newNodeStat(node Node) *nodeStat {
	return &nodeStat{
		node: node,
		defs: newNodeStatCell(node),
		refs: newNodeStatCell(node),
	}
}

type nodeStat struct {
	node Node
	defs *nodeStatCell
	refs *nodeStatCell
}

func (s *nodeStat) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"node": s.node,
		"defs": s.defs,
		"refs": s.refs,
	})
}
func (s *nodeStat) Node() Node         { return s.node }
func (s *nodeStat) Defs() NodeStatCell { return s.defs }
func (s *nodeStat) Refs() NodeStatCell { return s.refs }
func (s *nodeStat) Weight() int {
	var n int
	for _, x := range s.defs.Deps() {
		n += x.Weight()
	}
	for _, x := range s.refs.Deps() {
		n += x.Weight()
	}
	return n
}

func newNodeStatCell(node Node) *nodeStatCell {
	return &nodeStatCell{
		node: node,
		deps: map[string]*nodeStatDep{},
	}
}

type nodeStatCell struct {
	node Node
	deps map[string]*nodeStatDep
}

func (s *nodeStatCell) Weight() int {
	var n int
	for _, dep := range s.deps {
		n += dep.weight
	}
	return n
}
func (s *nodeStatCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"node": s.node,
		"deps": s.deps,
	})
}
func (s *nodeStatCell) add(node Node) {
	if x, ok := s.deps[node.ID()]; ok {
		x.weight++
		return
	}
	s.deps[node.ID()] = newNodeStatDep(node)
}
func (s *nodeStatCell) Node() Node { return s.node }
func (s *nodeStatCell) Deps() []NodeStatDep {
	var (
		deps = make([]NodeStatDep, len(s.deps))
		i    int
	)
	for _, x := range s.deps {
		deps[i] = x
		i++
	}
	return deps
}
func (s *nodeStatCell) Get(node Node) (NodeStatDep, bool) {
	dep, ok := s.deps[node.ID()]
	return dep, ok
}

func newNodeStatDep(node Node) *nodeStatDep {
	return &nodeStatDep{
		node:   node,
		weight: 1,
	}
}

type nodeStatDep struct {
	node   Node
	weight int
}

func (s *nodeStatDep) Node() Node  { return s.node }
func (s *nodeStatDep) Weight() int { return s.weight }
func (s *nodeStatDep) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"node":   s.node,
		"weight": s.weight,
	})
}

func NewNodeStatCalculator() NodeStatCalculator {
	return &nodeStatCalculator{
		d: map[string]*nodeStat{},
	}
}

type nodeStatCalculator struct {
	d map[string]*nodeStat
}

func (s *nodeStatCalculator) Add(ref, def Node) {
	if x, ok := s.d[ref.ID()]; ok {
		x.refs.add(def)
	} else {
		x := newNodeStat(ref)
		x.refs.add(def)
		s.d[ref.ID()] = x
	}
	if x, ok := s.d[def.ID()]; ok {
		x.defs.add(ref)
	} else {
		x := newNodeStat(def)
		x.defs.add(ref)
		s.d[def.ID()] = x
	}
}

func (s *nodeStatCalculator) Result() NodeStatSet {
	return &nodeStatSet{
		stats: s.d,
	}
}

/* node dependencies */

type (
	NodeDep interface {
		Ref() Node
		Def() Node
		Weight() int
	}

	nodeDep struct {
		ref    Node
		def    Node
		weight int
	}

	NodeDepCalculator interface {
		Add(ref, def Node)
		Result() []NodeDep
	}
)

func (s *nodeDep) Ref() Node   { return s.ref }
func (s *nodeDep) Def() Node   { return s.def }
func (s *nodeDep) Weight() int { return s.weight }
func (s *nodeDep) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"ref":    s.ref,
		"def":    s.def,
		"weight": s.weight,
	})
}

func NewNodeDepCalculator() NodeDepCalculator {
	return &nodeDepCalculator{
		d: map[string]*nodeDep{},
	}
}

type nodeDepCalculator struct {
	d map[string]*nodeDep
}

func (*nodeDepCalculator) id(ref, def Node) string {
	return fmt.Sprintf("%s>%s", ref.ID(), def.ID())
}

func (s *nodeDepCalculator) Add(ref, def Node) {
	id := s.id(ref, def)
	if dep, found := s.d[id]; found {
		dep.weight++
		return
	}
	s.d[id] = &nodeDep{
		ref:    ref,
		def:    def,
		weight: 1,
	}
}

func (s *nodeDepCalculator) Result() []NodeDep {
	var (
		i int
		r = make([]NodeDep, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}
