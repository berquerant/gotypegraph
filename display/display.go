package display

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/berquerant/gotypegraph/display/jsonify"
	"github.com/berquerant/gotypegraph/search"
)

type (
	Writer interface {
		Write(search.Use) error
		Flush() error
	}

	WriterConfig struct {
	}

	WriterOption func(*WriterConfig)
)

func NewJSONWriter(w io.Writer) Writer {
	return &jsonWriter{
		w: w,
	}
}

type jsonWriter struct {
	w io.Writer
}

func (s *jsonWriter) Write(node search.Use) error {
	if err := s.write(node); err != nil {
		return fmt.Errorf("JSONWriter: %w", err)
	}
	return nil
}

func (s *jsonWriter) write(node search.Use) error {
	js := jsonify.NewUse(node)
	b, err := json.Marshal(js)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.w, string(b))
	return err
}

func (*jsonWriter) Flush() error { return nil }
