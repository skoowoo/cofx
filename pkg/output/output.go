package output

import (
	"io"
	"os"
)

type OutputOption func(*Output)

type Output struct {
	w io.Writer
}

func New(opts ...OutputOption) *Output {
	out := &Output{
		w: os.Stdout,
	}
	for _, opt := range opts {
		opt(out)
	}
	return out
}

// Discard create a 'Output' object, use it to discard the output content
func Discard() *Output {
	return New(WithWriter(io.Discard))
}

// Stdout create a 'Output' object, use it to print the output content to stdout of the OS
func Stdout() *Output {
	return New(WithWriter(os.Stdout))
}

// File create a 'Output' object, use it to write the output content into a file, the parameter
// is filename
func File(name string) *Output {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	return New(WithWriter(f))
}

func (o *Output) Write(p []byte) (int, error) {
	return o.w.Write(p)
}

func WithWriter(w io.Writer) OutputOption {
	return func(o *Output) {
		o.w = w
	}
}
