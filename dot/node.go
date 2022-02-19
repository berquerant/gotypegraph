package dot

import (
	"fmt"
	"regexp"
)

var escapeTargetRegex = regexp.MustCompile(`[/$.()]`)

func Escape(v string) string {
	return string(escapeTargetRegex.ReplaceAll([]byte(v), []byte("_")))
}

type ID string

func (s ID) Raw() string    { return string(s) }
func (s ID) String() string { return Escape(string(s)) }

type (
	Node interface {
		ID() ID
		AttrList() AttrList
		String() string
	}

	NodeConfig struct {
		attrList AttrList
	}

	NodeOption func(*NodeConfig)

	node struct {
		id   ID
		conf *NodeConfig
	}
)

func NewNode(id ID, opt ...NodeOption) Node {
	var config NodeConfig
	for _, x := range opt {
		x(&config)
	}
	return &node{
		id:   id,
		conf: &config,
	}
}

func WithNodeAttrList(attrList AttrList) NodeOption {
	return func(c *NodeConfig) {
		c.attrList = attrList
	}
}

func (s *node) ID() ID             { return s.id }
func (s *node) AttrList() AttrList { return s.conf.attrList }
func (s *node) String() string {
	if s.conf.attrList == nil || s.conf.attrList.Len() == 0 {
		return fmt.Sprintf("%s;", s.id.String())
	}
	return fmt.Sprintf("%s %s;", s.id, s.conf.attrList.String(false))
}
