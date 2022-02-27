package testpkg

import (
	"fmt"

	"github.com/berquerant/gotypegraph/search/testpkg/sub"
)

const C1 = "C1"

var V1, V2 = "V1", func() string {
	return C1 + "X"
}()

type X struct {
	FX1 int
}

type Y struct {
	FY1 *X
}

func (*Y) SameNameFunc() string {
	return V1
}

func (x *X) SameNameFunc() {
	fmt.Println("in X.SameNameFunc", V1)
}

func SameNameFunc() {
	println("in SameNameFunc")
	sub.SameNameFunc()
}
