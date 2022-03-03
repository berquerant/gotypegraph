package search

import (
	"go/ast"
	"go/types"

	"github.com/berquerant/gotypegraph/logger"
	"golang.org/x/tools/go/packages"
)

type (
	// Target is the dependency of the object.
	Target interface {
		Ident() *ast.Ident // ref
		Obj() Object       // def
	}

	TargetExtractor interface {
		Extract(pkg *packages.Package, filter Filter) <-chan Target
	}

	TargetExtractorConfig struct {
		resultBufferSize int
	}

	TargetExtractorOption func(*TargetExtractorConfig)
)

func NewTargetExtractor(opt ...TargetExtractorOption) TargetExtractor {
	config := TargetExtractorConfig{
		resultBufferSize: 1000,
	}
	for _, x := range opt {
		x(&config)
	}
	return &targetExtractor{
		conf: &config,
	}
}

func WithTargetExtractorResultBufferSize(v int) TargetExtractorOption {
	return func(c *TargetExtractorConfig) {
		if v >= 0 {
			c.resultBufferSize = v
		}
	}
}

type targetExtractor struct {
	conf *TargetExtractorConfig
}

func (s *targetExtractor) Extract(pkg *packages.Package, filter Filter) <-chan Target {
	resultC := make(chan Target, s.conf.resultBufferSize)
	go func() {
		for ident, obj := range pkg.TypesInfo.Uses {
			logger.Debugf("[TargetExtractor] %s (%s) %s %s", pkg.Name, pkg.PkgPath, ident, types.ObjectString(obj, nil))
			tgt := NewTarget(ident, obj)
			if filter != nil && filter(tgt) {
				resultC <- tgt
			}
		}
		close(resultC)
	}()
	return resultC
}

func NewTarget(ident *ast.Ident, obj Object) Target {
	return &target{
		ident: ident,
		obj:   obj,
	}
}

type target struct {
	ident *ast.Ident
	obj   Object
}

func (s *target) Ident() *ast.Ident { return s.ident }
func (s *target) Obj() Object       { return s.obj }
