package logfile

import (
	"io"
	"os"
	"sync"
)

type LogfileOption func(*Logfile)

type Logfile struct {
	sync.Mutex
	w        io.WriteCloser
	filePath string
}

func New(opts ...LogfileOption) *Logfile {
	out := &Logfile{
		w: os.Stdout,
	}
	for _, opt := range opts {
		opt(out)
	}
	return out
}

func (o *Logfile) Write(p []byte) (int, error) {
	o.Lock()
	defer o.Unlock()

	return o.w.Write(p)
}

// TODO:
func (o *Logfile) Read(p []byte) (int, error) {
	return 0, nil
}

// TODO:
func (o *Logfile) ReadLine() {

}

func (o *Logfile) IsFile() bool {
	o.Lock()
	defer o.Unlock()
	return o.filePath != ""
}

func (o *Logfile) Reset() error {
	// The 'Reset' method is only valid for 'File' type, so Reset will close the file and then reopen it
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

func (o *Logfile) Close() error {
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
func Stdout() *Logfile {
	return New(WithWriter(os.Stdout))
}

// File create a 'Output' object, use it to write the output content into a file, the argument
// is filename
func File(name string) (*Logfile, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return New(WithWriter(f), WithFilePath(name)), nil
}

func WithWriter(w io.WriteCloser) LogfileOption {
	return func(o *Logfile) {
		o.w = w
	}
}

func WithFilePath(p string) LogfileOption {
	return func(o *Logfile) {
		o.filePath = p
	}
}
