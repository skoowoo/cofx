package output

import (
	"bytes"
	"io"
)

type Output struct {
	W          io.Writer
	HandleFunc func(line []byte)
	buffer     []byte
}

func (o *Output) Write(p []byte) (n int, err error) {
	l := len(p)
	for i := 0; i < l; {
		end := bytes.IndexByte(p[i:], '\n')
		if end != -1 {
			line := p[i : i+end+1]
			if len(o.buffer) > 0 {
				line = append(o.buffer, line...)
				o.buffer = nil
			}
			if o.HandleFunc != nil {
				o.HandleFunc(line)
			}
			i = i + end + 1
			continue
		}
		o.buffer = p[i:]
		break
	}
	if o.W != nil {
		return o.W.Write(p)
	} else {
		return l, nil
	}
}

func (o *Output) Close() {
	if len(o.buffer) > 0 && o.HandleFunc != nil {
		o.HandleFunc(o.buffer)
	}
	o.buffer = nil
}
