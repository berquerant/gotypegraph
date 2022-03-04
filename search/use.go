package search

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sync"

	"github.com/berquerant/gotypegraph/astutil"
	"github.com/berquerant/gotypegraph/logger"
	"github.com/berquerant/gotypegraph/util"
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

type (
	UseSearcherConfig struct {
		resultBufferSize  int
		workerNum         int
		searchUniverse    bool
		searchForeign     bool
		searchPrivate     bool
		ignorePkgSelfloop bool
		ignoreUseSelfloop bool
		pkgNameRegexp     util.RegexpPair
		objNameRegexp     util.RegexpPair
	}

	UseSearcherOption func(*UseSearcherConfig)

	UseSearcher interface {
		Search() <-chan Use
	}
)

func NewUseSearcher(
	pkgs []*packages.Package,
	refSearcher RefPkgSearcher,
	objExtractor ObjExtractor,
	tgtExtractor TargetExtractor,
	fieldSearcher FieldSearcher,
	defSetFilter Filter,
	opt ...UseSearcherOption,
) UseSearcher {
	config := UseSearcherConfig{
		resultBufferSize: 1000,
		workerNum:        4,
	}
	for _, x := range opt {
		x(&config)
	}

	pkgSet := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		logger.Verbosef("[UseSearcher] with pkg %s (%s)", pkg.Name, pkg.PkgPath)
		pkgSet[pkg.PkgPath] = pkg
	}
	filter := defSetFilter
	if !config.searchPrivate {
		logger.Debugf("[UseSearcher] use exported filter")
		filter = filter.And(ExportedFilter)
	}
	if config.searchForeign {
		logger.Debugf("[UseSearcher] use foreign filter")
		filter = filter.Or(OtherPkgFilter(pkgs))
	}
	if config.searchUniverse {
		logger.Debugf("[UseSearcher] use universe filter")
		filter = filter.Or(UniverseFilter)
	}
	if config.pkgNameRegexp != nil {
		logger.Debugf("[UseSearcher] use pkg name filter")
		filter = filter.And(PkgNameFilter(config.pkgNameRegexp))
	}
	if config.objNameRegexp != nil {
		logger.Debugf("[UseSearcher] use obj name filter")
		filter = filter.And(ObjectNameFilter(config.objNameRegexp))
	}
	if config.ignorePkgSelfloop {
		logger.Debugf("[UseSearcher] ignore pkg self loop")
	}
	if config.ignoreUseSelfloop {
		logger.Debugf("[UseSearcher] ignore use self loop")
	}

	return &useSearcher{
		pkgSet:        pkgSet,
		objExtractor:  objExtractor,
		refSearcher:   refSearcher,
		tgtExtractor:  tgtExtractor,
		fieldSearcher: fieldSearcher,
		filter:        filter,
		conf:          &config,
	}
}

type useSearcher struct {
	pkgSet        map[string]*packages.Package // pkg path => pkg
	objExtractor  ObjExtractor
	refSearcher   RefPkgSearcher
	tgtExtractor  TargetExtractor
	fieldSearcher FieldSearcher
	filter        Filter
	conf          *UseSearcherConfig
}

func (s *useSearcher) Search() <-chan Use {
	logger.Debugf("[UseSearcher] start search %d worker %d buffer", s.conf.workerNum, s.conf.resultBufferSize)
	defer logger.Debugf("[UseSearcher] end search")
	var (
		resultC = make(chan Use, s.conf.resultBufferSize)
		pkgC    = make(chan *packages.Package, s.conf.workerNum)
		wg      sync.WaitGroup
	)
	wg.Add(s.conf.workerNum)
	for i := 0; i < s.conf.workerNum; i++ {
		go func() {
			defer wg.Done()
			for pkg := range pkgC {
				s.search(pkg, resultC)
			}
		}()
	}
	go func() {
		for _, pkg := range s.pkgSet {
			pkgC <- pkg
		}
		close(pkgC)
		wg.Wait()
		close(resultC)
	}()
	return resultC
}

func (s *useSearcher) search(pkg *packages.Package, resultC chan<- Use) {
	if !s.selectPkg(pkg) {
		return
	}
	var targetNum int
	logger.Verbosef("[UseSearcher] search %s %s", pkg.Name, pkg.PkgPath)
	defer func() {
		logger.Verbosef("[UseSearcher] searched %s %s %d targets", pkg.Name, pkg.PkgPath, targetNum)
	}()

	for tgt := range s.tgtExtractor.Extract(pkg, s.filter) {
		targetNum++
		logger.Debugf("[UseSearcher] target %s (%s) %s %s",
			pkg.Name,
			pkg.PkgPath,
			tgt.Ident(),
			types.ObjectString(tgt.Obj().(types.Object), nil),
		)

		if s.ignorePkgSelfloop(pkg, tgt.Obj()) {
			continue
		}

		astNode, ok := s.refSearcher.Search(pkg, tgt.Ident().Pos())
		if !ok {
			// ref's ast should be found
			continue
		}
		var (
			valueSpecIndex = s.findValueSpecIndex(astNode, tgt.Ident().Pos())
			refObj         = s.findObj(pkg, astNode, valueSpecIndex)
		)

		if s.ignoreUseSelfloop(refObj, tgt.Obj()) {
			continue
		}

		rNode := NewRefNode(
			NewPkg(pkg),
			refObj,
			&NodeInfo{
				ValueSpecIndex: valueSpecIndex,
			},
			astNode,
			tgt.Ident(),
		)

		var (
			dNode DefNode
			recv  = s.findRecv(tgt.Obj().Pkg(), tgt.Obj().Pos())
		)
		switch {
		case tgt.Obj().Pkg() == nil:
			dNode = NewDefNode(
				NewBuiltinPkg(),
				tgt.Obj(),
				&NodeInfo{
					Recv: recv,
				},
			)
		default:
			objPkg := tgt.Obj().Pkg()
			if defPkg, ok := s.pkgSet[objPkg.Path()]; ok {
				dNode = NewDefNode(
					NewPkg(defPkg),
					tgt.Obj(),
					&NodeInfo{
						Recv: recv,
					},
				)
			} else {
				dNode = NewDefNode(
					NewPkgWithName(objPkg.Name(), objPkg.Path()),
					tgt.Obj(),
					&NodeInfo{
						Recv: recv,
					},
				)
			}
		}
		resultC <- NewUse(rNode, dNode)
	}
}

func (s *useSearcher) ignoreUseSelfloop(left, right Object) bool {
	if !s.conf.ignoreUseSelfloop {
		return false
	}
	return s.hasSameObj(left, right)
}

func (*useSearcher) hasSameObj(left, right Object) bool {
	if (left.Pkg() == nil) != (right.Pkg() == nil) {
		return false
	}
	return left.Pkg() == nil || (left.Pkg().Path() == right.Pkg().Path() && left.Pos() == right.Pos())
}

func (s *useSearcher) ignorePkgSelfloop(refPkg *packages.Package, defObj Object) bool {
	if !s.conf.ignorePkgSelfloop {
		return false
	}
	return s.hasSamePkg(refPkg, defObj)
}

func (*useSearcher) hasSamePkg(refPkg *packages.Package, defObj Object) bool {
	return defObj.Pkg() != nil && defObj.Pkg().Path() == refPkg.PkgPath
}

func (*useSearcher) findValueSpecIndex(node ast.Node, pos token.Pos) int {
	if vs, ok := node.(*ast.ValueSpec); ok {
		if idx, ok := astutil.FindValueSpecIndex(vs, pos); ok {
			return idx
		}
	}
	return -1
}

func (s *useSearcher) findRecv(pkg *types.Package, pos token.Pos) string {
	if typeName, ok := s.fieldSearcher.Search(pkg, pos); ok {
		return typeName.Name()
	}
	return ""
}

func (s *useSearcher) findObj(pkg *packages.Package, node ast.Node, valueSpecIndex int) Object {
	var ident *ast.Ident
	switch node := node.(type) {
	case *ast.FuncDecl:
		ident = node.Name
	case *ast.TypeSpec:
		ident = node.Name
	case *ast.ValueSpec:
		ident = node.Names[valueSpecIndex]
	default:
		return nil
	}
	if obj, ok := s.objExtractor.Extract(pkg, ident); ok {
		return obj
	}
	return nil
}

func (s *useSearcher) selectPkg(pkg *packages.Package) bool {
	return s.conf.pkgNameRegexp == nil || s.conf.pkgNameRegexp.MatchString(pkg.Name)
}

func WithUseSearcherIgnoreUseSelfloop(v bool) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.ignoreUseSelfloop = v
	}
}

func WithUseSearcherIgnorePkgSelfloop(v bool) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.ignorePkgSelfloop = v
	}
}

func WithUseSearcherObjNameRegexp(v util.RegexpPair) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.objNameRegexp = v
	}
}

func WithUseSearcherPkgNameRegexp(v util.RegexpPair) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.pkgNameRegexp = v
	}
}

func WithUseSearcherSearchPrivate(v bool) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.searchPrivate = v
	}
}

func WithUseSearcherSearchForeign(v bool) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.searchForeign = v
	}
}

func WithUseSearcherSearchUniverse(v bool) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.searchUniverse = v
	}
}

func WithUseSearcherWorkerNum(v int) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.workerNum = v
	}
}

func WithUseSearcherResultBufferSize(v int) UseSearcherOption {
	return func(c *UseSearcherConfig) {
		c.resultBufferSize = v
	}
}

type (
	Use interface {
		Ref() RefNode
		Def() DefNode
	}

	use struct {
		ref RefNode
		def DefNode
	}
)

func NewUse(ref RefNode, def DefNode) Use {
	return &use{
		ref: ref,
		def: def,
	}
}

func (s *use) Ref() RefNode { return s.ref }
func (s *use) Def() DefNode { return s.def }

type (
	Node interface {
		Pkg() Pkg
		Obj() Object
		Type() NodeType
		Name() string
		RecvString(opt ...NodeOption) string
		Info() *NodeInfo
	}

	NodeConfig struct {
		rawRecv bool
	}

	NodeOption func(*NodeConfig)

	RefNode interface {
		Node
		AST() ast.Node
		Ident() *ast.Ident
	}

	DefNode interface {
		Node
	}

	NodeInfo struct {
		// ValueSpecIndex is the index of the ValueSpec.Specs that correspond to Obj() if AST() is ValueSpec.
		// -1 is a invalid value.
		ValueSpecIndex int
		Recv           string
	}
)

func WithNodeRawRecv(v bool) NodeOption {
	return func(c *NodeConfig) {
		c.rawRecv = v
	}
}

type refNode struct {
	*node
	astNode ast.Node
	ident   *ast.Ident
}

func NewRefNode(pkg Pkg, obj Object, info *NodeInfo, astNode ast.Node, ident *ast.Ident) RefNode {
	return &refNode{
		node:    newNode(pkg, obj, info),
		astNode: astNode,
		ident:   ident,
	}
}

func (s *refNode) AST() ast.Node     { return s.astNode }
func (s *refNode) Ident() *ast.Ident { return s.ident }
func (s *refNode) String() string {
	var b util.StringBuilder
	b.Writef("%s %s (%d) %s %s",
		s.pkg,
		s.ident,
		s.ident.Pos(),
		s.nodeType,
		reflect.TypeOf(s.astNode),
	)
	if s.obj != nil {
		b.Write(" " + types.ObjectString(s.obj.(types.Object), nil))
	}
	return b.String()
}

type defNode struct {
	*node
}

func NewDefNode(pkg Pkg, obj Object, info *NodeInfo) DefNode {
	return &defNode{
		newNode(pkg, obj, info),
	}
}

func (s *defNode) String() string {
	return types.ObjectString(s.obj.(types.Object), nil)
}

type node struct {
	pkg      Pkg
	obj      Object
	nodeType NodeType
	nodeInfo *NodeInfo
}

func newNode(pkg Pkg, obj Object, info *NodeInfo) *node {
	return &node{
		pkg:      pkg,
		obj:      obj,
		nodeType: NewNodeType(obj),
		nodeInfo: info,
	}
}

func (s *node) Pkg() Pkg        { return s.pkg }
func (s *node) Obj() Object     { return s.obj }
func (s *node) Type() NodeType  { return s.nodeType }
func (s *node) Info() *NodeInfo { return s.nodeInfo }
func (s *node) Name() string    { return s.obj.Name() }
func (s *node) RecvString(opt ...NodeOption) string {
	var conf NodeConfig
	for _, x := range opt {
		x(&conf)
	}

	if recv := s.nodeInfo.Recv; recv != "" {
		return recv
	}

	sig, ok := s.obj.Type().(*types.Signature)
	if !ok {
		return ""
	}
	if sig.Recv() == nil {
		return ""
	}
	switch t := sig.Recv().Type().(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Pointer:
		if nmd, ok := t.Elem().(*types.Named); ok {
			if conf.rawRecv {
				return nmd.Obj().Name()
			}
			return "*" + nmd.Obj().Name()
		}
		return ""
	default:
		return ""
	}
}

type (
	Pkg interface {
		Pkg() *packages.Package
		Name() string
		Path() string
		IsBuiltin() bool
	}

	builtinPkg struct{}
	pkgWithPkg struct {
		pkg *packages.Package
	}
	pkgWithName struct {
		name string
		path string
	}
)

const builtinPkgName = "builtin"

func NewBuiltinPkg() Pkg { return &builtinPkg{} }

func (*builtinPkg) Pkg() *packages.Package { return nil }
func (*builtinPkg) Name() string           { return builtinPkgName }
func (*builtinPkg) Path() string           { return builtinPkgName }
func (*builtinPkg) IsBuiltin() bool        { return true }
func (*builtinPkg) String() string         { return builtinPkgName }

func NewPkg(pkg *packages.Package) Pkg {
	return &pkgWithPkg{
		pkg: pkg,
	}
}

func (s *pkgWithPkg) Pkg() *packages.Package { return s.pkg }
func (s *pkgWithPkg) Name() string           { return s.pkg.Name }
func (s *pkgWithPkg) Path() string           { return s.pkg.PkgPath }
func (*pkgWithPkg) IsBuiltin() bool          { return false }
func (s *pkgWithPkg) String() string {
	return fmt.Sprintf("%s path %s id %s", s.pkg.Name, s.pkg.PkgPath, s.pkg.ID)
}

func NewPkgWithName(name, path string) Pkg {
	return &pkgWithName{
		name: name,
		path: path,
	}
}

func (*pkgWithName) Pkg() *packages.Package { return nil }
func (s *pkgWithName) Name() string         { return s.name }
func (s *pkgWithName) Path() string         { return s.path }
func (*pkgWithName) IsBuiltin() bool        { return false }
func (s *pkgWithName) String() string {
	return fmt.Sprintf("name %s path %s", s.name, s.path)
}

type NodeType int

func NewNodeType(obj Object) NodeType {
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

const (
	UnknownNodeType NodeType = iota
	BuiltinNodeType
	FuncNodeType
	MethodNodeType
	TypeNodeType
	VarNodeType
	ConstNodeType
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
