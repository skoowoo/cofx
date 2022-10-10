package pretty

import (
	"strings"

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

	s := lipgloss.JoinVertical(lipgloss.Left, ts.Render(title), ds.Render(desc))
	b.result = s
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

// TableBlock used to render a table as a block.
type TableBlock struct {
	result string
}

func NewTableBlock(titles []string, values [][]string) TableBlock {
	// Calculate the max width of each column.
	rowCount := len(values)
	colCount := len(values[0])
	widths := make([]int, colCount)
	for c := 0; c < colCount; c++ {
		for r := 0; r < rowCount; r++ {
			s := values[r][c]
			if w := lipgloss.Width(s); w > widths[c] {
				widths[c] = w
			}
		}
	}
	for i, t := range titles {
		if w := lipgloss.Width(t); w > widths[i] {
			widths[i] = w
		}
	}

	// Create style for each column.
	styles := make(map[int]lipgloss.Style)
	for i, w := range widths {
		if w > 2 {
			w += 2
		}
		styles[i] = lipgloss.NewStyle().Width(w)
	}

	var buf strings.Builder
	// Render the table header.
	for i, t := range titles {
		s := styles[i].Render(ColorGrey1.Render(t))
		buf.WriteString(s)
	}
	buf.WriteString("\n")
	// Render the table body.
	for _, row := range values {
		for i, col := range row {
			s := styles[i].Render(col)
			buf.WriteString(s)
		}
		buf.WriteString("\n")
	}
	return TableBlock{
		result: buf.String(),
	}
}

func (tb TableBlock) Render() string {
	return tb.result
}

func (tb TableBlock) Width() int {
	return lipgloss.Width(tb.result)
}

func (tb TableBlock) Height() int {
	return lipgloss.Height(tb.result)
}

// TextBlock used to render a text line as a block
type TextBlock struct {
	result string
}

func NewTextBlock(s string) TextBlock {
	return TextBlock{s}
}

func (tx TextBlock) Render() string {
	return tx.result
}

func (tx TextBlock) Height() int {
	return lipgloss.Height(tx.result)
}

func (tx TextBlock) Width() int {
	return lipgloss.Width(tx.result)
}
