package logfile

import (
	"errors"
	"io"
	"os"
	"sync"
)

type LogfileOption func(*Logfile)

type Logfile struct {
	sync.Mutex
	w        io.Writer
	filePath string
	file     *os.File
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

func WithWriter(w io.WriteCloser) LogfileOption {
	return func(o *Logfile) {
		o.w = w
	}
}

func WithFile(p string, f *os.File) LogfileOption {
	return func(o *Logfile) {
		o.filePath = p
		o.file = f
	}
}

// Write implements the io.Writer interface
func (o *Logfile) Write(p []byte) (int, error) {
	o.Lock()
	defer o.Unlock()

	return o.w.Write(p)
}

// Read implements the io.Reader interface
func (o *Logfile) Read(p []byte) (int, error) {
	if !o.IsFile() {
		return 0, errors.New("not a file")
	}
	return o.file.Read(p)
}

func (o *Logfile) IsFile() bool {
	o.Lock()
	defer o.Unlock()
	return o.filePath != "" && o.file != nil
}

// Reset close the file and then reopen it with truncate mode, will clear the content of the file
// Reset method is only available for 'file' type.
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
	o.file = f
	o.Unlock()

	return nil
}

// Close close the file if the 'Logfile' object is a file type
func (o *Logfile) Close() error {
	o.Lock()
	defer o.Unlock()

	if o.filePath != "" && o.file != nil {
		o.file.Close()
	}
	o.w = nil
	return nil
}

// Stdout create a 'Output' object, use it to print the output content to stdout of the OS
func Stdout() *Logfile {
	return New(WithWriter(os.Stdout))
}

// TruncFile create a 'Logfile' object, use it to write the output content into a file, the argument
// is filename; if the file exists, it will be truncated.
func TruncFile(name string) (*Logfile, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return New(WithWriter(f), WithFile(name, f)), nil
}

// File createa 'Logfile' object, can use it to read the ouput content from the file, the arugment is
// filename.
func File(name string) (*Logfile, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return New(WithWriter(f), WithFile(name, f)), nil
}
