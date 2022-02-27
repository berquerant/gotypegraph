package stat_test

import (
	"encoding/json"
	"testing"

	"github.com/berquerant/gotypegraph/search"
	"github.com/berquerant/gotypegraph/stat"
	"github.com/stretchr/testify/assert"
)

type mockNode struct {
	id string
}

func (s *mockNode) ID() string      { return s.id }
func (*mockNode) Pkg() stat.Pkg     { return nil }
func (*mockNode) Node() search.Node { return nil }

func TestNodeStatCalculator(t *testing.T) {
	const wantJSON = `{"p1":{"defs":{"deps":{"p1":{"node":{},"weight":1}},"node":{}},"node":{},"refs":{"deps":{"p1":{"node":{},"weight":1},"p2":{"node":{},"weight":1}},"node":{}}},"p2":{"defs":{"deps":{"p1":{"node":{},"weight":1},"p3":{"node":{},"weight":1}},"node":{}},"node":{},"refs":{"deps":{"p3":{"node":{},"weight":2}},"node":{}}},"p3":{"defs":{"deps":{"p2":{"node":{},"weight":2}},"node":{}},"node":{},"refs":{"deps":{"p2":{"node":{},"weight":1}},"node":{}}}}`
	// p1.defs: (p1, 1)
	// p1.refs: (p1, 1), (p2, 1)
	// p2.defs: (p1, 2), (p3, 1)
	// p2.refs: (p3, 2)
	// p3.defs: (p2, 2)
	// p3.refs: (p2, 1)

	c := stat.NewNodeStatCalculator()
	for _, x := range []struct {
		r string
		d string
	}{
		{r: "p1", d: "p1"}, // self loop
		{r: "p1", d: "p2"},
		{r: "p2", d: "p3"},
		{r: "p2", d: "p3"}, // weight 2
		{r: "p3", d: "p2"},
	} {
		c.Add(&mockNode{
			id: x.r,
		}, &mockNode{
			id: x.d,
		})
	}

	var (
		want interface{}
		got  interface{}
	)
	gotJSON, err := json.Marshal(c.Result())
	assert.Nil(t, err)
	t.Logf("%s", gotJSON)
	assert.Nil(t, json.Unmarshal([]byte(wantJSON), &want))
	assert.Nil(t, json.Unmarshal(gotJSON, &got))
	assert.Equal(t, want, got)
}

func TestNodeDepCalculator(t *testing.T) {
	want := map[string]map[string]int{
		"p1": {
			"p1": 1,
			"p2": 1,
		},
		"p2": {
			"p3": 2,
		},
		"p3": {
			"p2": 1,
		},
	}

	wantLen := 4
	c := stat.NewNodeDepCalculator()
	for _, x := range []struct {
		r string
		d string
	}{
		{r: "p1", d: "p1"}, // self loop
		{r: "p1", d: "p2"},
		{r: "p2", d: "p3"},
		{r: "p2", d: "p3"}, // weight 2
		{r: "p3", d: "p2"},
	} {
		c.Add(&mockNode{
			id: x.r,
		}, &mockNode{
			id: x.d,
		})
	}
	got := c.Result()
	assert.Equal(t, wantLen, len(got))
	for _, g := range got {
		dst, ok := want[g.Ref().ID()]
		assert.True(t, ok)
		w, ok := dst[g.Def().ID()]
		assert.True(t, ok)
		assert.Equal(t, w, g.Weight())
	}
}
