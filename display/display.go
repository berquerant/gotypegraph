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
		minFontsize int
		maxFontsize int
		minPenwidth int
		maxPenwidth int
		minWeight   int
		maxWeight   int
	}

	WriterOption func(*WriterConfig)
)

/* rankings */

const (
	minFontsize = 8
	maxFontsize = 24
	minPenwidth = 1
	maxPenwidth = 5
	minWeight   = 1
	maxWeight   = 100
)

func newWriterConfig() *WriterConfig {
	return &WriterConfig{
		minFontsize: minFontsize,
		maxFontsize: maxFontsize,
		minPenwidth: minPenwidth,
		maxPenwidth: maxPenwidth,
		minWeight:   minWeight,
		maxWeight:   maxWeight,
	}
}

func WithWriterMinFontsize(v int) WriterOption {
	return func(c *WriterConfig) {
		c.minFontsize = v
	}
}

func WithWriterMaxFontsize(v int) WriterOption {
	return func(c *WriterConfig) {
		c.maxFontsize = v
	}
}

func WithWriterMinPenwidth(v int) WriterOption {
	return func(c *WriterConfig) {
		c.minPenwidth = v
	}
}

func WithWriterMaxPenwidth(v int) WriterOption {
	return func(c *WriterConfig) {
		c.maxPenwidth = v
	}
}

func WithWriterMinWeight(v int) WriterOption {
	return func(c *WriterConfig) {
		c.minWeight = v
	}
}

func WithWriterMaxWeight(v int) WriterOption {
	return func(c *WriterConfig) {
		c.maxWeight = v
	}
}

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
