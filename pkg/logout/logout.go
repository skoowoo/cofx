package logout

import (
	"io"
	"os"
	"sync"
)

type OutputOption func(*Output)

type Output struct {
	sync.Mutex
	w        io.WriteCloser
	filePath string
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

func (o *Output) Write(p []byte) (int, error) {
	o.Lock()
	defer o.Unlock()

	return o.w.Write(p)
}

func (o *Output) IsFile() bool {
	o.Lock()
	defer o.Unlock()
	return o.filePath != ""
}

func (o *Output) Reset() error {
	if !o.IsFile() {
		return nil
	}
	if err := o.Close(); err != nil {
		return err
	}
	f, err := os.OpenFile(o.filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	o.Lock()
	o.w = f
	o.Unlock()

	return nil
}

func (o *Output) Close() error {
	o.Lock()
	defer o.Unlock()

	w := o.w
	o.w = nil
	if o.filePath != "" {
		return w.Close()
	}
	return nil
}

// Stdout create a 'Output' object, use it to print the output content to stdout of the OS
func Stdout() *Output {
	return New(WithWriter(os.Stdout))
}

// File create a 'Output' object, use it to write the output content into a file, the argument
// is filename
func File(name string) (*Output, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return New(WithWriter(f), WithFilePath(name)), nil
}

func WithWriter(w io.WriteCloser) OutputOption {
	return func(o *Output) {
		o.w = w
	}
}

func WithFilePath(p string) OutputOption {
	return func(o *Output) {
		o.filePath = p
	}
}
