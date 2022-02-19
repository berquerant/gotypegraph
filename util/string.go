package util

import (
	"fmt"
	"strings"
)

type StringBuilder struct {
	strings.Builder
}

func (s *StringBuilder) Write(v string) error {
	_, err := s.WriteString(v)
	return err
}

func (s *StringBuilder) Writeln(v string) error {
	_, err := s.WriteString(v + "\n")
	return err
}

func (s *StringBuilder) Writef(format string, v ...interface{}) error {
	_, err := s.WriteString(fmt.Sprintf(format, v...))
	return err
}

func (s *StringBuilder) Writelnf(format string, v ...interface{}) error {
	_, err := s.WriteString(fmt.Sprintf(format, v...) + "\n")
	return err
}
