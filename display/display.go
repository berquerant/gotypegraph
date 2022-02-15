package display

import (
	"fmt"
	"go/types"
	"io"
	"strings"

	"github.com/berquerant/gotypegraph/use"
)

type Writer interface {
	Write(result *use.Result) error
	Flush() error
}

func NewStringWriter(w io.Writer) Writer {
	return &stringWriter{
		w: w,
	}
}

type stringWriter struct {
	w io.Writer
}

func (s *stringWriter) Write(result *use.Result) error {
	if err := s.write(result); err != nil {
		return fmt.Errorf("StringWriter: %w", err)
	}
	return nil
}

func (s *stringWriter) write(result *use.Result) error {
	var (
		b strings.Builder
		r = result.RefPair
		d = result.DefPair
		w = func(format string, v ...interface{}) error {
			_, err := b.WriteString(fmt.Sprintf(format, v...))
			return err
		}
		recv = func(rv string) string {
			if rv == "" {
				return ""
			}
			return fmt.Sprintf("(%s).", rv)
		}
	)
	if err := w("[%s] %s%s:%s (%s) %s | ",
		r.Pkg.Name,
		recv(r.Recv()),
		r.NodeName(),
		r.NodeType(),
		r.Ident,
		r.Pkg.Fset.Position(r.Ident.Pos()),
	); err != nil {
		return err
	}
	if err := w("[%s] %s%s:%s %s",
		d.PkgName,
		recv(d.Recv()),
		d.NodeName(),
		d.NodeType(),
		types.ObjectString(d.Obj.(types.Object), nil), // NOTE: implements types.Object when returned from Searcher.Search().
	); err != nil {
		return err
	}
	if d.Pkg != nil {
		if err := w(" %s", d.Pkg.Fset.Position(d.Obj.Pos())); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(s.w, b.String())
	return err
}

func (*stringWriter) Flush() error { return nil }
