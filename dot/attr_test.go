package dot_test

import (
	"testing"

	"github.com/berquerant/gotypegraph/dot"
	"github.com/stretchr/testify/assert"
)

func TestAttr(t *testing.T) {
	for _, tc := range []struct {
		title       string
		attr        dot.Attr
		asStatement bool
		want        string
	}{
		{
			title: "attr",
			attr:  dot.NewAttr("akey", "aval"),
			want:  `akey="aval"`,
		},
		{
			title: "raw attr",
			attr:  dot.NewAttr("akey", "aval", dot.WithAttrRaw(true)),
			want:  "akey=aval",
		},
		{
			title:       "stmt",
			attr:        dot.NewAttr("akey", "aval"),
			asStatement: true,
			want:        `akey="aval";`,
		},
		{
			title:       "raw stmt",
			attr:        dot.NewAttr("akey", "aval", dot.WithAttrRaw(true)),
			asStatement: true,
			want:        `akey=aval;`,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.attr.String(tc.asStatement))
		})
	}
}

type mockAttr struct {
	str    string
	asStmt bool
}

func (*mockAttr) Key() string   { return "" }
func (*mockAttr) Value() string { return "" }
func (*mockAttr) IsRaw() bool   { return false }
func (s *mockAttr) String(asStmt bool) string {
	s.asStmt = asStmt
	return s.str
}

func TestAttrList(t *testing.T) {
	for _, tc := range []struct {
		title           string
		attrList        []string
		asStatementList bool
		want            string
	}{
		{
			title: "no attrs",
		},
		{
			title:    "an attr",
			attrList: []string{"astr"},
			want:     "[astr]",
		},
		{
			title:           "an attr stmt",
			attrList:        []string{"astr"},
			asStatementList: true,
			want:            "astr",
		},
		{
			title:    "attrs",
			attrList: []string{"astr", "bstr"},
			want:     "[astr,bstr]",
		},
		{
			title:           "attrs stmt",
			attrList:        []string{"astr", "bstr"},
			asStatementList: true,
			want: `astr
bstr`,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			list := dot.NewAttrList()
			attrList := make([]*mockAttr, len(tc.attrList))
			for i, a := range tc.attrList {
				attr := &mockAttr{
					str: a,
				}
				attrList[i] = attr
				list.Add(attr)
			}
			assert.Equal(t, tc.want, list.String(tc.asStatementList))
			for _, a := range attrList {
				assert.Equal(t, tc.asStatementList, a.asStmt)
			}
		})
	}
}
