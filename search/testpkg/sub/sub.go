package sub

import (
	"fmt"
	"os"
)

func SameNameFunc() {
	fmt.Fprintln(os.Stderr, "in sub.SameNameFunc")
}
