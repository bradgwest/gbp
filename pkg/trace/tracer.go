package trace

import (
	"fmt"
	"io"
)

// Tracer is the interface that describes an object capable of tracing events
// through code.
type Tracer interface {
	Trace(...interface{})
}

// New creates a new Tracer
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type tracer struct {
	out io.Writer
}

func (t tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}

type nilTracer struct{}

func (t *nilTracer) Trace(...interface{}) {}

// Off returns a nil tracer
func Off() Tracer {
	return &nilTracer{}
}
