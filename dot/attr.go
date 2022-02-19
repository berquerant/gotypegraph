package dot

import (
	"fmt"
	"strings"
)

type (
	Attr interface {
		Key() string
		Value() string
		IsRaw() bool
		String(asStatement bool) string
	}

	AttrConfig struct {
		isRaw bool
	}

	AttrOption func(*AttrConfig)

	attr struct {
		key   string
		value string
		conf  *AttrConfig
	}
)

func NewAttr(key, value string, opt ...AttrOption) Attr {
	var config AttrConfig
	for _, x := range opt {
		x(&config)
	}
	return &attr{
		key:   key,
		value: value,
		conf:  &config,
	}
}

func (s *attr) Key() string   { return s.key }
func (s *attr) Value() string { return s.value }
func (s *attr) IsRaw() bool   { return s.conf.isRaw }

func (s *attr) getValue() string {
	if s.conf.isRaw {
		return s.value
	}
	return fmt.Sprintf(`"%s"`, s.value)
}

func (s *attr) String(asStatement bool) string {
	if asStatement {
		return fmt.Sprintf("%s=%s;", s.key, s.getValue())
	}
	return fmt.Sprintf("%s=%s", s.key, s.getValue())
}

func WithAttrRaw(v bool) AttrOption {
	return func(c *AttrConfig) {
		c.isRaw = v
	}
}

type (
	AttrList interface {
		Len() int
		Add(Attr) AttrList
		Slice() []Attr
		String(asStatementList bool) string
	}

	attrList struct {
		list []Attr
	}
)

func NewAttrList() AttrList {
	return &attrList{}
}

func (s *attrList) Len() int { return len(s.list) }
func (s *attrList) Add(a Attr) AttrList {
	s.list = append(s.list, a)
	return s
}
func (s *attrList) Slice() []Attr { return s.list }
func (s *attrList) String(asStatementList bool) string {
	if len(s.list) == 0 {
		return ""
	}
	ss := make([]string, len(s.list))
	for i, a := range s.list {
		ss[i] = a.String(asStatementList)
	}
	if asStatementList {
		return strings.Join(ss, "\n")
	}
	return fmt.Sprintf("[%s]", strings.Join(ss, ","))
}
