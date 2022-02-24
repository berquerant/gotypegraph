package stat_test

import (
	"encoding/json"
	"testing"

	"github.com/berquerant/gotypegraph/search"
	"github.com/berquerant/gotypegraph/stat"
	"github.com/stretchr/testify/assert"
)

type mockPkg struct {
	id string
}

func (s *mockPkg) ID() string    { return s.id }
func (*mockPkg) Pkg() search.Pkg { return nil }

func TestPkgStatCalculator(t *testing.T) {
	const wantJSON = `{"p1":{"defs":{"deps":{"p1":{"pkg":"p1","weight":1}},"pkg":"p1"},"pkg":"p1","refs":{"deps":{"p1":{"pkg":"p1","weight":1},"p2":{"pkg":"p2","weight":1}},"pkg":"p1"},"weight":3},"p2":{"defs":{"deps":{"p1":{"pkg":"p1","weight":1},"p3":{"pkg":"p3","weight":1}},"pkg":"p2"},"pkg":"p2","refs":{"deps":{"p3":{"pkg":"p3","weight":2}},"pkg":"p2"},"weight":4},"p3":{"defs":{"deps":{"p2":{"pkg":"p2","weight":2}},"pkg":"p3"},"pkg":"p3","refs":{"deps":{"p2":{"pkg":"p2","weight":1}},"pkg":"p3"},"weight":3}}`
	// p1.defs: (p1, 1)
	// p1.refs: (p1, 1), (p2, 1)
	// p2.defs: (p1, 2), (p3, 1)
	// p2.refs: (p3, 2)
	// p3.defs: (p2, 2)
	// p3.refs: (p2, 1)

	c := stat.NewPkgStatCalculator()
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
		c.Add(&mockPkg{
			id: x.r,
		}, &mockPkg{
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

func TestPkgDepCalculator(t *testing.T) {
	c := stat.NewPkgDepCalculator()
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
		c.Add(&mockPkg{
			id: x.r,
		}, &mockPkg{
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
