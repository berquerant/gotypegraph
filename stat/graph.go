package stat

import (
	"fmt"
	"go/token"
	"path/filepath"

	"github.com/berquerant/gotypegraph/use"
	"github.com/berquerant/gotypegraph/util"
)

type Graph interface {
	Add(pair *Pair)
	PkgToNodes() map[string][]*NodeWeight
	Deps() []*Dep
	PkgDeps() []*PkgDep
	Pkgs() []*PkgWeight
	PkgIOs() map[string]*PkgIO
	NodeIOs() *NodeIOSet
}

func NewGraph() Graph {
	return &graph{
		pairs: []*Pair{},
	}
}

type graph struct {
	pairs []*Pair
}

func (s *graph) Add(pair *Pair) {
	s.pairs = append(s.pairs, pair)
}

func (s *graph) Deps() []*Dep {
	set := newDepSet()
	for _, pair := range s.pairs {
		set.add(pair)
	}
	return set.slice()
}

func (s *graph) Pkgs() []*PkgWeight {
	set := newPkgWeightSet()
	for _, pair := range s.pairs {
		set.add(pair.Ref)
		set.add(pair.Def)
	}
	return set.slice()
}

func (s *graph) PkgDeps() []*PkgDep {
	set := newPkgDepSet()
	for _, pair := range s.pairs {
		set.add(pair)
	}
	return set.slice()
}

func (s *graph) PkgToNodes() map[string][]*NodeWeight {
	set := newNodeWeightSet()
	for _, pair := range s.pairs {
		set.add(pair.Ref)
		set.add(pair.Def)
	}
	d := map[string][]*NodeWeight{}
	for _, x := range set.slice() {
		d[x.Node.Pkg] = append(d[x.Node.Pkg], x)
	}
	return d
}

func (s *graph) PkgIOs() map[string]*PkgIO {
	set := newPkgIOSet()
	for _, pair := range s.pairs {
		set.add(pair)
	}
	return set.d
}

func (s *graph) NodeIOs() *NodeIOSet {
	set := newNodeIOSet()
	for _, pair := range s.pairs {
		set.add(pair)
	}
	return set
}

/* basic data structures */

type (
	// Node is a symbol.
	Node struct {
		Pkg      string
		Name     string
		Position token.Position
		Type     use.NodeType
		// Recv is the receiver type name iff the node is a method.
		Recv string
	}
	// Pair is a dependency.
	Pair struct {
		Ref *Node
		Def *Node
	}
)

const unknownPositionFilename = "unknown"

func newUnknownPosition() token.Position {
	return token.Position{
		Filename: unknownPositionFilename,
	}
}

func positionToDir(pos token.Position) string {
	if pos.Filename == unknownPositionFilename {
		return "unknown"
	}
	return filepath.Dir(pos.Filename)
}

func NewPair(result *use.Result) *Pair {
	var (
		r = result.RefPair
		d = result.DefPair
	)
	// TODO: add struct name that field belongs
	return &Pair{
		Ref: &Node{
			Pkg:      r.Pkg.Name,
			Name:     r.NodeName(),
			Position: r.Pkg.Fset.Position(r.Node.Pos()),
			Type:     r.NodeType(),
			Recv:     r.Recv(),
		},
		Def: &Node{
			Pkg:  d.PkgName,
			Name: d.NodeName(),
			Position: func() token.Position {
				if d.Pkg == nil {
					return newUnknownPosition()
				}
				return d.Pkg.Fset.Position(d.Obj.Pos())
			}(),
			Type: d.NodeType(),
			Recv: d.Recv(),
		},
	}
}

func (s *Node) id() string    { return fmt.Sprintf("%s.%s", s.Pkg, s.Name) }
func (s *Pair) id() string    { return fmt.Sprintf("%s>%s", s.Ref.id(), s.Def.id()) }
func (s *Pair) pkgID() string { return fmt.Sprintf("%s>%s", s.Ref.Pkg, s.Def.Pkg) }

/* result data types */

type (
	// Dep is a dependency with frequency of appearance.
	Dep struct {
		Pair   *Pair
		Weight int
	}
	// PkgDep is a package level dependency.
	PkgDep struct {
		Ref    string
		Def    string
		Weight int
	}
	// NodeWeight is a node level frequency of appearance.
	NodeWeight struct {
		Node   *Node
		Weight int
	}
	// PkgWeight is a package level frequency of appearance.
	PkgWeight struct {
		Pkg      string
		Weight   int
		Position string
	}

	NodeIO struct {
		Node       *Node
		In         int
		Out        int
		inUniqSet  *util.StringSet
		outUniqSet *util.StringSet
	}

	PkgIO struct {
		Pkg        string
		In         int
		Out        int
		Position   string
		inUniqSet  *util.StringSet
		outUniqSet *util.StringSet
	}
)

func (s *NodeIO) UniqIn() int  { return s.inUniqSet.Len() }
func (s *NodeIO) UniqOut() int { return s.outUniqSet.Len() }

func (s *PkgIO) UniqIn() int  { return s.inUniqSet.Len() }
func (s *PkgIO) UniqOut() int { return s.outUniqSet.Len() }

type (
	NodeIOSet struct {
		d map[string]*NodeIO
	}
	pkgIOSet struct {
		d map[string]*PkgIO
	}
)

/* stat calculators */

func newPkgIOSet() *pkgIOSet {
	return &pkgIOSet{
		d: map[string]*PkgIO{},
	}
}

func (s *pkgIOSet) add(pair *Pair) {
	var (
		r = pair.Ref
		d = pair.Def
	)
	if x, found := s.d[r.Pkg]; found {
		x.Out++
		x.outUniqSet.Add(d.Pkg)
	} else {
		s.d[r.Pkg] = &PkgIO{ // FIXME: more unique key
			Pkg:        r.Pkg,
			Out:        1,
			In:         0,
			Position:   positionToDir(r.Position),
			outUniqSet: util.NewStringSet().Add(d.Pkg),
			inUniqSet:  util.NewStringSet(),
		}
	}
	if x, found := s.d[d.Pkg]; found {
		x.In++
		x.inUniqSet.Add(r.Pkg)
	} else {
		s.d[d.Pkg] = &PkgIO{ // FIXME: more unique key
			Pkg:        d.Pkg,
			Out:        0,
			In:         1,
			Position:   positionToDir(d.Position),
			outUniqSet: util.NewStringSet(),
			inUniqSet:  util.NewStringSet().Add(r.Pkg),
		}
	}
}

func newNodeIOSet() *NodeIOSet {
	return &NodeIOSet{
		d: map[string]*NodeIO{},
	}
}

func (s *NodeIOSet) Get(node *Node) (*NodeIO, bool) {
	x, ok := s.d[node.id()]
	return x, ok
}

func (s *NodeIOSet) add(pair *Pair) {
	var (
		r = pair.Ref
		d = pair.Def
	)
	if x, found := s.d[r.id()]; found {
		x.Out++
		x.outUniqSet.Add(d.id())
	} else {
		s.d[r.id()] = &NodeIO{
			Node:       r,
			Out:        1,
			In:         0,
			outUniqSet: util.NewStringSet().Add(r.id()),
			inUniqSet:  util.NewStringSet(),
		}
	}
	if x, found := s.d[d.id()]; found {
		x.In++
		x.inUniqSet.Add(r.id())
	} else {
		s.d[d.id()] = &NodeIO{
			Node:       d,
			Out:        0,
			In:         1,
			outUniqSet: util.NewStringSet(),
			inUniqSet:  util.NewStringSet().Add(r.id()),
		}
	}
}

type (
	nodeWeightSet struct {
		d map[string]*NodeWeight
	}
	depSet struct {
		d map[string]*Dep
	}
	pkgDepSet struct {
		d map[string]*PkgDep
	}
	pkgWeightSet struct {
		d map[string]*PkgWeight
	}
)

func newPkgWeightSet() *pkgWeightSet {
	return &pkgWeightSet{
		d: map[string]*PkgWeight{},
	}
}

func (s *pkgWeightSet) add(node *Node) {
	if n, ok := s.d[node.Pkg]; ok {
		n.Weight++
		return
	}
	s.d[node.Pkg] = &PkgWeight{ // FIXME: more unique key
		Pkg:      node.Pkg,
		Weight:   1,
		Position: positionToDir(node.Position),
	}
}

func (s *pkgWeightSet) slice() []*PkgWeight {
	var (
		i int
		r = make([]*PkgWeight, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}

func newNodeWeightSet() *nodeWeightSet {
	return &nodeWeightSet{
		d: map[string]*NodeWeight{},
	}
}

func (s *nodeWeightSet) add(node *Node) {
	if n, ok := s.d[node.id()]; ok {
		n.Weight++
		return
	}
	s.d[node.id()] = &NodeWeight{
		Node:   node,
		Weight: 1,
	}
}

func (s *nodeWeightSet) slice() []*NodeWeight {
	var (
		i int
		r = make([]*NodeWeight, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}

func newDepSet() *depSet {
	return &depSet{
		d: map[string]*Dep{},
	}
}

func (s *depSet) add(pair *Pair) {
	if d, ok := s.d[pair.id()]; ok {
		d.Weight++
		return
	}
	s.d[pair.id()] = &Dep{
		Pair:   pair,
		Weight: 1,
	}
}

func (s *depSet) slice() []*Dep {
	var (
		i int
		r = make([]*Dep, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}

func newPkgDepSet() *pkgDepSet {
	return &pkgDepSet{
		d: map[string]*PkgDep{},
	}
}

func (s *pkgDepSet) add(pair *Pair) {
	if d, ok := s.d[pair.pkgID()]; ok {
		d.Weight++
		return
	}
	s.d[pair.pkgID()] = &PkgDep{ // FIXME: more unique key
		Ref:    pair.Ref.Pkg,
		Def:    pair.Def.Pkg,
		Weight: 1,
	}
}

func (s *pkgDepSet) slice() []*PkgDep {
	var (
		i int
		r = make([]*PkgDep, len(s.d))
	)
	for _, x := range s.d {
		r[i] = x
		i++
	}
	return r
}
