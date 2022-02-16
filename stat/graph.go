package stat

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/berquerant/gotypegraph/use"
	"github.com/berquerant/gotypegraph/util"
	"golang.org/x/tools/go/packages"
)

type Graph interface {
	Add(pair *Pair)
	PkgToNodes() map[PkgKey][]*NodeWeight
	Deps() []*Dep
	PkgDeps() []*PkgDep
	Pkgs() []*PkgWeight
	PkgIOs() map[PkgKey]*PkgIO
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

func (s *graph) PkgToNodes() map[PkgKey][]*NodeWeight {
	set := newNodeWeightSet()
	for _, pair := range s.pairs {
		set.add(pair.Ref)
		set.add(pair.Def)
	}
	d := map[PkgKey][]*NodeWeight{}
	for _, x := range set.slice() {
		d[x.Node.PkgKey] = append(d[x.Node.PkgKey], x)
	}
	return d
}

func (s *graph) PkgIOs() map[PkgKey]*PkgIO {
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
		Name     string
		Position token.Position
		Type     use.NodeType
		// Recv is the receiver type name iff the node is a method.
		Recv   string
		PkgKey PkgKey
	}
	// Pair is a dependency.
	Pair struct {
		Ref *Node
		Def *Node
	}
	// PkgKey identifies a package by package name and package dir.
	PkgKey struct {
		pkgName string
		pkgDir  string
	}
)

// NewPkgKey returns a new PkgKey.
func NewPkgKey(pkg *packages.Package, pos token.Pos) PkgKey {
	if pkg == nil {
		return PkgKey{}
	}
	return PkgKey{
		pkgName: pkg.Name,
		pkgDir:  filepath.Dir(pkg.Fset.Position(pos).Filename),
	}
}

const unknownPkgKey = "unknown-package-key"

func (s *PkgKey) LString() string {
	if s == nil {
		return unknownPkgKey
	}
	if s.pkgDir == "" {
		return s.pkgName
	}
	xs := strings.Split(s.pkgDir, "/")
	return fmt.Sprintf("%s/%s", xs[len(xs)-1], s.pkgName)
}

func (s *PkgKey) String() string {
	if s == nil {
		return unknownPkgKey
	}
	if s.pkgDir == "" {
		return s.pkgName
	}
	return fmt.Sprintf("%s/%s", s.pkgDir, s.pkgName)
}

// Key returns a string for unique dot node id.
func (s *PkgKey) Key() string {
	if s == nil {
		return unknownPkgKey
	}
	if s.pkgDir == "" {
		return s.pkgName
	}
	return fmt.Sprintf("%s__%s", s.pkgName, util.AsDotID(s.pkgDir))
}

func (s *PkgKey) Pkg() string {
	if s == nil {
		return ""
	}
	return s.pkgName
}

func (s *PkgKey) Dir() string {
	if s == nil {
		return ""
	}
	return s.pkgDir
}

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
			Name:     r.NodeName(),
			Position: r.Pkg.Fset.Position(r.Node.Pos()),
			Type:     r.NodeType(),
			Recv:     r.Recv(),
			PkgKey:   NewPkgKey(r.Pkg, r.Node.Pos()),
		},
		Def: &Node{
			Name: d.NodeName(),
			Position: func() token.Position {
				if d.Pkg == nil {
					return newUnknownPosition()
				}
				return d.Pkg.Fset.Position(d.Obj.Pos())
			}(),
			Type: d.NodeType(),
			Recv: d.Recv(),
			PkgKey: func() PkgKey {
				if d.Pkg == nil {
					return PkgKey{
						pkgName: d.PkgName,
					}
				}
				return NewPkgKey(d.Pkg, d.Obj.Pos())
			}(),
		},
	}
}

func (s *Node) id() string    { return fmt.Sprintf("%s.%s", s.PkgKey.Key(), s.Name) }
func (s *Pair) id() string    { return fmt.Sprintf("%s>%s", s.Ref.id(), s.Def.id()) }
func (s *Pair) pkgID() string { return fmt.Sprintf("%s>%s", s.Ref.PkgKey.Key(), s.Def.PkgKey.Key()) }

/* result data types */

type (
	// Dep is a dependency with frequency of appearance.
	Dep struct {
		Pair   *Pair
		Weight int
	}
	// PkgDep is a package level dependency.
	PkgDep struct {
		Ref    PkgKey
		Def    PkgKey
		Weight int
	}
	// NodeWeight is a node level frequency of appearance.
	NodeWeight struct {
		Node   *Node
		Weight int
	}
	// PkgWeight is a package level frequency of appearance.
	PkgWeight struct {
		PkgKey   PkgKey
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
		PkgKey     PkgKey
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
		d map[PkgKey]*PkgIO
	}
)

/* stat calculators */

func newPkgIOSet() *pkgIOSet {
	return &pkgIOSet{
		d: map[PkgKey]*PkgIO{},
	}
}

func (s *pkgIOSet) add(pair *Pair) {
	var (
		r = pair.Ref
		d = pair.Def
	)
	if x, found := s.d[r.PkgKey]; found {
		x.Out++
		x.outUniqSet.Add(r.PkgKey.Key())
	} else {
		s.d[r.PkgKey] = &PkgIO{
			PkgKey:     r.PkgKey,
			Out:        1,
			In:         0,
			Position:   positionToDir(r.Position),
			outUniqSet: util.NewStringSet().Add(d.PkgKey.Key()),
			inUniqSet:  util.NewStringSet(),
		}
	}
	if x, found := s.d[d.PkgKey]; found {
		x.In++
		x.inUniqSet.Add(r.PkgKey.Key())
	} else {
		s.d[d.PkgKey] = &PkgIO{
			PkgKey:     d.PkgKey,
			Out:        0,
			In:         1,
			Position:   positionToDir(d.Position),
			outUniqSet: util.NewStringSet(),
			inUniqSet:  util.NewStringSet().Add(r.PkgKey.Key()),
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
		d map[PkgKey]*PkgWeight
	}
)

func newPkgWeightSet() *pkgWeightSet {
	return &pkgWeightSet{
		d: map[PkgKey]*PkgWeight{},
	}
}

func (s *pkgWeightSet) add(node *Node) {
	if n, ok := s.d[node.PkgKey]; ok {
		n.Weight++
		return
	}
	s.d[node.PkgKey] = &PkgWeight{
		PkgKey:   node.PkgKey,
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
	s.d[pair.pkgID()] = &PkgDep{
		Ref:    pair.Ref.PkgKey,
		Def:    pair.Def.PkgKey,
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
