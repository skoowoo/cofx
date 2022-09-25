package logset

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/cofxlabs/cofx/pkg/output"
	"github.com/cofxlabs/cofx/pkg/uidesign"
)

type LogsetOption func(*Logset)

func WithAddr(addr string) LogsetOption {
	return func(ls *Logset) {
		ls.addr = addr
		ls.typ = "File"
	}
}

func WithStdout() LogsetOption {
	return func(ls *Logset) {
		ls.addr = "Stdout"
		ls.typ = "Stdout"
	}
}

func New(opts ...LogsetOption) *Logset {
	ls := &Logset{
		buckets: make(map[string]*LogBucket),
	}
	for _, opt := range opts {
		opt(ls)
	}
	return ls
}

// Logset be used to log as a log service.
type Logset struct {
	sync.Mutex
	addr    string
	typ     string
	buckets map[string]*LogBucket
}

// Restore restore all buckets from the log directory.
func (s *Logset) Restore() error {
	s.Lock()
	defer s.Unlock()
	if s.typ == "File" {
		err := filepath.Walk(s.addr, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("%w: access path '%s'", err, path)
			}
			if info.IsDir() {
				id := info.Name()
				s.buckets[id] = &LogBucket{
					id:      id,
					set:     s,
					writers: make(map[string]interface{}),
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateBucket create a new bucket object that can be used to write the output content.
func (s *Logset) CreateBucket(bucketid string) *LogBucket {
	s.Lock()
	defer s.Unlock()

	if b, ok := s.buckets[bucketid]; ok {
		return b
	}
	bucket := &LogBucket{
		id:      bucketid,
		set:     s,
		writers: make(map[string]interface{}),
	}
	s.buckets[bucketid] = bucket
	return bucket
}

// GetBucket returns the bucket object by bucketid.
func (s *Logset) GetBucket(bucketid string) (*LogBucket, error) {
	s.Lock()
	defer s.Unlock()

	if bucket, ok := s.buckets[bucketid]; ok {
		return bucket, nil
	}
	return nil, errors.New("bucket not found: " + bucketid)
}

type LogBucket struct {
	id      string
	set     *Logset
	writers map[string]interface{}
}

func (b *LogBucket) IsFile() bool {
	return b.set.typ == "File"
}

func (b *LogBucket) IsStdout() bool {
	return b.set.typ == "Stdout"
}

func (b *LogBucket) Reset() error {
	for _, w := range b.writers {
		if lf, ok := w.(*logFile); ok {
			if err := lf.Reset(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *LogBucket) CreateWriter(id string, ws ...io.Writer) (io.Writer, error) {
	if b.IsFile() {
		path := filepath.Join(b.set.addr, "buckets", b.id, id, "logfile")
		lf, err := newLogFile2Write(path)
		if err != nil {
			return nil, fmt.Errorf("%w: create writer", err)
		}
		if _, ok := b.writers[id]; ok {
			return nil, errors.New("writer already exists: " + id)
		}
		b.writers[id] = lf
		return lf, nil
	}
	if b.IsStdout() {
		var w io.Writer = os.Stdout
		if len(ws) > 0 {
			w = ws[0]
		}
		lout := newLogStdout(id, w)
		if _, ok := b.writers[id]; ok {
			return nil, errors.New("writer already exists: " + id)
		}
		b.writers[id] = lout
		return lout, nil
	}
	return nil, nil
}

func (b *LogBucket) CreateReader(id string) (io.ReadCloser, error) {
	if b.IsFile() {
		path := filepath.Join(b.set.addr, "buckets", b.id, id, "logfile")
		lf, err := newLogFile2Read(path)
		if err != nil {
			return nil, fmt.Errorf("%w: create reader", err)
		}
		return lf, nil
	}
	if b.IsStdout() {
		return nil, errors.New("stdout can not create reader")
	}
	return nil, nil
}

type logFile struct {
	sync.Mutex
	file     *os.File
	filePath string
}

// newLogFile2Write create a 'LogFile' object, use it to write the output content into a file, the argument
// is filename; if the file exists, it will be truncated.
func newLogFile2Write(path string) (*logFile, error) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &logFile{
		file:     f,
		filePath: path,
	}, nil
}

// newLogFile2Read create 'Logfile' object, can use it to read the output content from the file, the argument is
// file path.
func newLogFile2Read(path string) (*logFile, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &logFile{
		file:     f,
		filePath: path,
	}, nil
}

// Write implements the io.Writer interface
func (lf *logFile) Write(p []byte) (int, error) {
	lf.Lock()
	defer lf.Unlock()
	return lf.file.Write(p)
}

// Read implements the io.Reader interface
func (lf *logFile) Read(p []byte) (int, error) {
	lf.Lock()
	defer lf.Unlock()
	return lf.file.Read(p)
}

// Close close the file if the 'Logfile' object is a file type
func (lf *logFile) Close() error {
	lf.Lock()
	defer lf.Unlock()

	if lf.file != nil {
		lf.file.Close()
	}
	lf.file = nil
	return nil
}

// Reset close the file and then reopen it with truncate mode, will clear the content of the file
// Reset method is only available for 'file' type.
func (lf *logFile) Reset() error {
	if err := lf.Close(); err != nil {
		return err
	}
	f, err := os.OpenFile(lf.filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	lf.Lock()
	lf.file = f
	lf.Unlock()

	return nil
}

type logStdout struct {
	sync.Mutex
	w   io.Writer
	out *output.Output
	// Usually, the id is the seq of this function
	id string
}

func newLogStdout(id string, w io.Writer) *logStdout {
	l := &logStdout{
		w:  w,
		id: id,
	}
	l.out = &output.Output{
		W: nil,
		HandleFunc: func(line []byte) {
			l.w.Write([]byte("  "))
			l.w.Write(line)
		},
	}
	return l
}

// Reset reset the 'logStdout' object, will clear the 'printedTitle' flag
func (l *logStdout) Reset() error {
	l.Lock()
	defer l.Unlock()
	return nil
}

func (l *logStdout) WriteTitle(primary, secondary string) {
	s := uidesign.IconCycle.String() + uidesign.ColorGrey.Render(primary+" ➜ "+secondary) + "\n"
	fmt.Fprint(l.w, s)
}

func (l *logStdout) WriteSummary(lines []string) {
	for _, s := range lines {
		fmt.Fprint(l.w, uidesign.IconRight.String()+s+"\n")
	}
	fmt.Fprint(l.w, " \n") // add a blank line
}

func (l *logStdout) Write(p []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.out.Write(p)
}
