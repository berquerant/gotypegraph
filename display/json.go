package display

import (
	"encoding/json"
	"fmt"
	"go/token"
	"io"

	"github.com/berquerant/gotypegraph/use"
)

func NewJSONWriter(w io.Writer) Writer {
	return &jsonWriter{
		w: w,
	}
}

type jsonWriter struct {
	w io.Writer
}

func (s *jsonWriter) Write(result *use.Result) error {
	if err := s.write(result); err != nil {
		return fmt.Errorf("JSONWriter: %w", err)
	}
	return nil
}

func (s *jsonWriter) write(result *use.Result) error {
	b, err := json.Marshal(NewResult(result))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.w, string(b))
	return err
}

func (*jsonWriter) Flush() error { return nil }

type (
	Pos struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	}
	RefPair struct {
		Pkg  string `json:"pkg"`
		Type string `json:"type"`
		Node string `json:"node"`
		Name string `json:"name"`
		Pos  *Pos   `json:"pos"`
		Recv string `json:"recv,omitempty"`
	}
	DefPair struct {
		Pkg  string `json:"pkg"`
		Type string `json:"type"`
		Name string `json:"name"`
		Pos  *Pos   `json:"pos,omitempty"`
		Recv string `json:"recv,omitempty"`
	}
	Result struct {
		RefPair *RefPair `json:"ref"`
		DefPair *DefPair `json:"def"`
	}
)

func NewPos(pos token.Position) *Pos {
	return &Pos{
		File:   pos.Filename,
		Line:   pos.Line,
		Column: pos.Column,
	}
}

func NewResult(result *use.Result) *Result {
	var (
		r  = result.RefPair
		d  = result.DefPair
		rp = &RefPair{
			Pkg:  r.Pkg.Name,
			Type: r.NodeType().String(),
			Node: r.NodeName(),
			Name: r.Ident.String(),
			Pos:  NewPos(r.Pkg.Fset.Position(r.Ident.Pos())),
			Recv: r.Recv(),
		}
		dp = &DefPair{
			Pkg:  d.PkgName,
			Type: d.NodeType().String(),
			Name: d.NodeName(),
			Recv: d.Recv(),
		}
	)
	if d.Pkg != nil {
		dp.Pos = NewPos(d.Pkg.Fset.Position(d.Obj.Pos()))
	}
	return &Result{
		RefPair: rp,
		DefPair: dp,
	}
}
