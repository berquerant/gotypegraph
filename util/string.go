package util

import (
	"fmt"
	"strings"
)

// StringBuilder wraps strings.Builder.
type StringBuilder struct {
	strings.Builder
}

func (s *StringBuilder) Writeln(v string) {
	_, _ = s.WriteString(v + "\n")
}

func (s *StringBuilder) Writelnf(format string, v ...interface{}) {
	_, _ = s.WriteString(fmt.Sprintf(format, v...) + "\n")
}
