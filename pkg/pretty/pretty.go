package pretty

import (
	"io"
)

type Pretty struct {
	w        io.Writer
	disabled bool
}

func (p *Pretty) SetDisabled() {
	p.disabled = true
}

type ListEl struct {
	Title   string
	Content []string
}

type List struct {
	els []ListEl
	Pretty
}

func NewList() *List {
	return &List{}
}

func (l *List) Append(title string, content []string) *List {
	l.els = append(l.els, ListEl{
		Title:   title,
		Content: content,
	})
	return l
}

func (l *List) Reset() {
	l.els = l.els[0:0]
}

func (l *List) Println() {

}

func (l *List) Fprintln() {

}
