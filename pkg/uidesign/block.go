package uidesign

import (
	"github.com/charmbracelet/lipgloss"
)

// Blocker define a block component in the window.
type Blocker interface {
	Render() string
	Width() int
	Height() int
}

// TitleBlock used to render a title block.
type TitleBlock struct {
	result string
}

func NewTitleBlock(title, desc string) TitleBlock {
	b := TitleBlock{}
	width := max(lipgloss.Width(title), lipgloss.Width(desc))

	ts := lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Italic(true).
		Foreground(lipgloss.Color("15"))
	ds := ColorGrey1.Copy().Width(width)

	b.result = lipgloss.JoinVertical(lipgloss.Left, ds.Render(title), ts.Render(desc))
	return b
}

func (t TitleBlock) Height() int {
	return lipgloss.Height(t.result)
}

func (t TitleBlock) Width() int {
	return lipgloss.Width(t.result)
}

func (t TitleBlock) Render() string {
	return t.result
}

// FooterBlock used to render some messages into the footer, e.g. help message.
type FooterBlock struct {
	style  lipgloss.Style
	result string
}

func NewFooterBlock(msg string) FooterBlock {
	b := FooterBlock{
		style: ColorGrey3.Copy(),
	}
	b.result = b.style.Render(msg)
	return b
}

func (f FooterBlock) Height() int {
	return lipgloss.Height(f.result)
}

func (f FooterBlock) Width() int {
	return lipgloss.Width(f.result)
}

func (f FooterBlock) Render() string {
	return f.result
}

// KvsBlock used to render key-value pairs as a block.
type KvsBlock struct {
	kstyle lipgloss.Style
	vstyle lipgloss.Style
	result string
}

func NewKvsBlock(kvs ...[]string) KvsBlock {
	b := KvsBlock{
		kstyle: ColorGrey1.Copy(),
		vstyle: lipgloss.NewStyle(),
	}
	var buf []string
	for _, kv := range kvs {
		if len(kv) != 2 {
			continue
		}
		k := kv[0]
		v := kv[1]

		s := lipgloss.JoinVertical(lipgloss.Left, b.vstyle.Render(v), b.kstyle.Render(k))
		s = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("99")).
			Align(lipgloss.Left).
			MarginLeft(0).MarginRight(1).
			Width(lipgloss.Width(s)).Render(s)

		buf = append(buf, s)
	}
	b.result = lipgloss.JoinHorizontal(lipgloss.Bottom, buf...)
	return b
}

func (kb KvsBlock) Height() int {
	return lipgloss.Width(kb.result)
}

func (kb KvsBlock) Width() int {
	return lipgloss.Height(kb.result)
}

func (kb KvsBlock) Render() string {
	return kb.result
}
