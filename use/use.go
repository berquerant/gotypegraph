package use

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"
	"sync"

	"github.com/berquerant/gotypegraph/def"
	"github.com/berquerant/gotypegraph/ref"
	"golang.org/x/tools/go/packages"
)

// Object is the set of the public methods of types.Object.
type Object interface {
	Parent() *types.Scope
	Pos() token.Pos
	Pkg() *types.Package
	Name() string
	Type() types.Type
	Exported() bool
	Id() string
}

// Filter selects objects.
type Filter func(*ast.Ident, Object) bool

func (s Filter) And(next Filter) Filter {
	if s == nil {
		return next
	}
	if next == nil {
		return s
	}
	return func(ident *ast.Ident, obj Object) bool {
		return s(ident, obj) && next(ident, obj)
	}
}

func (s Filter) Or(next Filter) Filter {
	if s == nil {
		return next
	}
	if next == nil {
		return s
	}
	return func(ident *ast.Ident, obj Object) bool {
		return s(ident, obj) || next(ident, obj)
	}
}

// Targets is a ref-def pair.
// Ident is ref, Obj is def.
type Target struct {
	Ident *ast.Ident
	Obj   types.Object
}

const extractTargetsBufferSize = 1000

// ExtractTargets extracts ref-def pairs from uses map.
func ExtractTargets(pkg *packages.Package, filter Filter) <-chan *Target {
	resultC := make(chan *Target, extractTargetsBufferSize)
	go func() {
		for ident, obj := range pkg.TypesInfo.Uses {
			if filter != nil && !filter(ident, obj) {
				continue
			}
			resultC <- &Target{
				Ident: ident,
				Obj:   obj,
			}
		}
		close(resultC)
	}()
	return resultC
}

// UniverseFilter selects objects in the universe scope.
func UniverseFilter(_ *ast.Ident, obj Object) bool { return obj != nil && obj.Pkg() == nil }

// ObjectNameFilter selects objects whose name matches accept but not deny.
func ObjectNameFilter(accept, deny *regexp.Regexp) Filter {
	return func(_ *ast.Ident, obj Object) bool {
		return obj != nil && (accept == nil || accept.MatchString(obj.Name())) &&
			(deny == nil || !deny.MatchString(obj.Name()))
	}
}

// PkgNameFilter selects objects whose package matches accept but not deny.
func PkgNameFilter(accept, deny *regexp.Regexp) Filter {
	return func(_ *ast.Ident, obj Object) bool {
		if obj == nil || obj.Pkg() == nil {
			return false
		}
		pkg := obj.Pkg().Name()
		return (accept == nil || accept.MatchString(pkg)) && (deny == nil || !deny.MatchString(pkg))
	}
}

// ExportedFilter selects exported objects.
func ExportedFilter(_ *ast.Ident, obj Object) bool { return obj != nil && obj.Exported() }

// OtherPkgFilter returns a filter function that selects objects not contained in pkgs.
func OtherPkgFilter(pkgs []*packages.Package) Filter {
	pkgSet := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		pkgSet[pkg.Name] = true // FIXME: more unique key
	}
	return func(_ *ast.Ident, obj Object) bool {
		return obj != nil && obj.Pkg() != nil && !pkgSet[obj.Pkg().Name()]
	}
}

// DefFilter returns a filter function that selects objects contained in the def set.
func DefFilter(sets []*def.Set) Filter {
	pkgSet := make(map[string]map[token.Pos]bool, len(sets))
	for _, set := range sets {
		posSet := map[token.Pos]bool{}
		for _, d := range set.Defs {
			// register the pos of the ident
			for _, vs := range d.ValueSpecs {
				for _, v := range vs.Names {
					posSet[v.Pos()] = true
				}
			}
			for _, fd := range d.FuncDecls {
				posSet[fd.Name.Pos()] = true
			}
			for _, ts := range d.TypeSpecs {
				posSet[ts.Pos()] = true
			}
		}
		pkgSet[set.Pkg.Name] = posSet // FIXME: more unique key
	}
	return func(_ *ast.Ident, obj Object) bool {
		if obj == nil || obj.Pkg() == nil {
			return false
		}
		p, ok := pkgSet[obj.Pkg().Name()]
		return ok && p[obj.Pos()]
	}
}

// Config is a configuration for Searcher.
type Config struct {
	searchForeign   bool
	searchUniverse  bool
	searchPrivate   bool
	acceptPkgRegex  *regexp.Regexp
	denyPkgRegex    *regexp.Regexp
	acceptNameRegex *regexp.Regexp
	denyNameRegex   *regexp.Regexp
}

type Option func(*Config)

// WithSearchForeign switches whether searches definitions in foreign packages.
func WithSearchForeign(v bool) Option {
	return func(c *Config) {
		c.searchForeign = v
	}
}

// WithSearchUniverse switches whether searches definitions in builtin package.
func WithSearchUniverse(v bool) Option {
	return func(c *Config) {
		c.searchUniverse = v
	}
}

// WithSearchPrivate switches whether searches package private definitions.
func WithSearchPrivate(v bool) Option {
	return func(c *Config) {
		c.searchPrivate = v
	}
}

// WithAcceptPkgRegex sets regex that selects packages to be searched.
func WithAcceptPkgRegex(r *regexp.Regexp) Option {
	return func(c *Config) {
		c.acceptPkgRegex = r
	}
}

// WithDenyPkgRegex sets regex that selects packages to be not searched.
func WithDenyPkgRegex(r *regexp.Regexp) Option {
	return func(c *Config) {
		c.denyPkgRegex = r
	}
}

// WithAcceptNameRegex sets regex that selects object name to be searched.
func WithAcceptNameRegex(r *regexp.Regexp) Option {
	return func(c *Config) {
		c.acceptNameRegex = r
	}
}

// WithDenyNameRegex sets regex that selects object name to be not searched.
func WithDenyNameRegex(r *regexp.Regexp) Option {
	return func(c *Config) {
		c.denyNameRegex = r
	}
}

func NewSearcher(pkgs []*packages.Package, refSearcher ref.Searcher, defFilter Filter, opts ...Option) Searcher {
	var config Config
	for _, x := range opts {
		x(&config)
	}
	pkgSet := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgSet[pkg.Name] = pkg // FIXME: more unique key
	}
	filter := defFilter
	if !config.searchPrivate {
		filter = filter.And(ExportedFilter)
	}
	if config.searchForeign {
		filter = filter.Or(OtherPkgFilter(pkgs))
	}
	if config.searchUniverse {
		filter = filter.Or(UniverseFilter)
	}
	if config.acceptPkgRegex != nil || config.denyPkgRegex != nil {
		filter = filter.And(PkgNameFilter(config.acceptPkgRegex, config.denyPkgRegex))
	}
	if config.acceptNameRegex != nil || config.denyNameRegex != nil {
		filter = filter.And(ObjectNameFilter(config.acceptPkgRegex, config.denyPkgRegex))
	}
	return &searcher{
		config:      &config,
		pkgs:        pkgSet,
		refSearcher: refSearcher,
		filter:      filter,
	}
}

type Searcher interface {
	// Search searches ref-def pairs.
	Search() <-chan *Result
}

type searcher struct {
	config      *Config
	pkgs        map[string]*packages.Package
	refSearcher ref.Searcher
	filter      Filter
}

const (
	searchResultBufferSize = 1000
	searchWorkerNum        = 4
)

func (s *searcher) Search() <-chan *Result {
	resultC := make(chan *Result, searchResultBufferSize)
	go func() {
		var (
			wg   sync.WaitGroup
			pkgC = make(chan *packages.Package, searchWorkerNum)
		)
		wg.Add(searchWorkerNum)
		for i := 0; i < searchWorkerNum; i++ {
			go func() {
				s.searchWorker(pkgC, resultC)
				wg.Done()
			}()
		}
		for _, pkg := range s.pkgs {
			if !s.selectPkg(pkg.Name) {
				continue
			}
			pkgC <- pkg
		}
		close(pkgC)
		wg.Wait()
		close(resultC)
	}()
	return resultC
}

func (s *searcher) selectPkg(pkg string) bool {
	var (
		accept = s.config.acceptPkgRegex
		deny   = s.config.denyPkgRegex
	)
	return (accept == nil || accept.MatchString(pkg)) && (deny == nil || !deny.MatchString(pkg))
}

func (s *searcher) searchWorker(pkgC <-chan *packages.Package, resultC chan<- *Result) {
	for pkg := range pkgC {
		s.search(pkg, resultC)
	}
}

func (s *searcher) search(pkg *packages.Package, resultC chan<- *Result) {
	for tgt := range ExtractTargets(pkg, s.filter) {
		node, found := s.refSearcher.Search(pkg, tgt.Ident.Pos())
		if !found {
			continue
		}
		rp := &RefPair{
			Pkg:   pkg,
			Node:  node,
			Ident: tgt.Ident,
		}
		switch {
		case tgt.Obj.Pkg() == nil:
			resultC <- &Result{
				RefPair: rp,
				DefPair: &DefPair{
					PkgName: "builtin",
					Obj:     tgt.Obj,
				},
			}
		default:
			if defPkg, found := s.pkgs[tgt.Obj.Pkg().Name()]; found {
				resultC <- &Result{
					RefPair: rp,
					DefPair: &DefPair{
						PkgName: tgt.Obj.Pkg().Name(),
						Pkg:     defPkg,
						Obj:     tgt.Obj,
					},
				}
			} else {
				resultC <- &Result{
					RefPair: rp,
					DefPair: &DefPair{
						PkgName: tgt.Obj.Pkg().Name(),
						Obj:     tgt.Obj,
					},
				}
			}
		}
	}
}

type (
	RefPair struct {
		Pkg   *packages.Package
		Node  ast.Node
		Ident *ast.Ident
	}
	DefPair struct {
		Pkg     *packages.Package // may be nil, nil means builtin
		PkgName string
		Obj     Object // NOTE: implements types.Object when returned from Searcher.Search().
	}
	Result struct {
		RefPair *RefPair
		DefPair *DefPair
	}
)

func (s *DefPair) NodeType() NodeType { return NewNodeTypeFromObj(s.Obj) }
func (s *DefPair) NodeName() string   { return s.Obj.Name() }

func (s *DefPair) Recv() string {
	if s.NodeType() != MethodNodeType {
		return ""
	}
	switch t := s.Obj.Type().(*types.Signature).Recv().Type().(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Pointer:
		if nmd, ok := t.Elem().(*types.Named); ok {
			return "*" + nmd.Obj().Name()
		}
		return ""
	default:
		return ""
	}
}

func (s *RefPair) Recv() string {
	if s.NodeType() != MethodNodeType {
		return ""
	}
	switch t := s.Node.(*ast.FuncDecl).Recv.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
		return ""
	default:
		return ""
	}
}

func (s *RefPair) NodeType() NodeType { return NewNodeTypeFromNode(s.Node) }

func (s *RefPair) NodeName() string {
	switch node := s.Node.(type) {
	case *ast.FuncDecl:
		return node.Name.String()
	case *ast.TypeSpec:
		return node.Name.String()
	case *ast.ValueSpec: // TODO: separate by name
		r := make([]string, len(node.Names))
		for i, x := range node.Names {
			r[i] = x.Name
		}
		return strings.Join(r, ",")
	default:
		return "UNKNOWN"
	}
}

type NodeType int

const (
	UnknownNodeType NodeType = iota
	// BuiltinNodeType means builtin definitions.
	BuiltinNodeType
	// FuncNodeType means functions.
	FuncNodeType
	// MethodNodeType means methods.
	MethodNodeType
	// TypeNodeType means type definitions.
	TypeNodeType
	// VarNodeType means variable definitions.
	VarNodeType
	// ConstNodeType means const definitions.
	ConstNodeType
	// FieldNodeType means struct field definitions.
	FieldNodeType
)

func (s NodeType) String() string {
	switch s {
	case BuiltinNodeType:
		return "builtin"
	case FuncNodeType:
		return "func"
	case MethodNodeType:
		return "method"
	case TypeNodeType:
		return "type"
	case VarNodeType:
		return "var"
	case ConstNodeType:
		return "const"
	case FieldNodeType:
		return "field"
	default:
		return "unknown"
	}
}

func NewNodeTypeFromNode(node ast.Node) NodeType {
	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Recv != nil {
			return MethodNodeType
		}
		return FuncNodeType
	case *ast.TypeSpec:
		return TypeNodeType
	case *ast.ValueSpec:
		for _, nm := range t.Names {
			if nm.Obj == nil || nm.Obj.Kind == ast.Var {
				return VarNodeType
			}
		}
		return ConstNodeType
	default:
		return UnknownNodeType
	}
}

func NewNodeTypeFromObj(obj Object) NodeType {
	switch obj := obj.(type) {
	case *types.Builtin:
		return BuiltinNodeType
	case *types.Const:
		return ConstNodeType
	case *types.Var:
		if obj.IsField() {
			return FieldNodeType
		}
		return VarNodeType
	case *types.TypeName:
		return TypeNodeType
	case *types.Func:
		if obj.Type().(*types.Signature).Recv() != nil {
			return MethodNodeType
		}
		return FuncNodeType
	default:
		return UnknownNodeType
	}
}
