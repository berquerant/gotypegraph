package load

import "golang.org/x/tools/go/packages"

type Loader interface {
	Load(patterns ...string) ([]*packages.Package, error)
}

func New() Loader { return &loader{} }

type loader struct{}

const loadMode = packages.NeedTypesInfo | packages.NeedTypes | packages.NeedName | packages.NeedSyntax | packages.NeedImports

func (*loader) Load(patterns ...string) ([]*packages.Package, error) {
	return packages.Load(&packages.Config{
		Mode: loadMode,
	}, patterns...)
}
