package search

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type (
	Target interface {
		Ident() *ast.Ident
		Obj() Object
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
