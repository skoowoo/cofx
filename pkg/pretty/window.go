package pretty

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Window is a grid of blocks, it indicates a canvas in the screen.
type Window struct {
	// Height is the avaliable height of the window.
	Height int
	// Width is the avaliable width of the window.
	Width int
	// Title describes title of the window.
	Title Row
	// Grid stores all blocks.
	Grid []Row
	// Footer can be used to show help message or others.
	Footer Row
	// Style defines the style of the window.
	Style lipgloss.Style
	// IsAltScreen indicates whether the window is in the alternate screen.
	IsAltScreen bool
}

// NewWindow create a new window instance.
func NewWindow(h, w int, alt bool) *Window {
	win := &Window{
		Height:      h,
		Width:       w,
		Style:       lipgloss.NewStyle().Width(w),
		Grid:        make([]Row, 1),
		IsAltScreen: alt,
	}
	if alt {
		win.Style.Align(lipgloss.Center)
	} else {
		win.Style.Align(lipgloss.Left)
	}
	return win
}

// SetTitle sets the title for the window.
func (w *Window) SetTitle(b Blocker) {
	w.Title = append(w.Title, b)
}

// SetFooter sets the footer for the window.
func (w *Window) SetFooter(b Blocker) {
	w.Footer = append(w.Footer, b)
}

// AppendBlock append a new block component into the window.
func (w *Window) AppendBlock(b Blocker) (int, int) {
	row := w.LastRow()
	if row.Width()+b.Width() > w.Width {
		w.Grid = append(w.Grid, Row{b})
		return len(w.Grid) - 1, 0
	}
	w.Grid[len(w.Grid)-1] = append(row, b)
	return len(w.Grid) - 1, len(w.LastRow()) - 1
}

// AppendNewRow append a new line into the window.
func (w *Window) AppendNewRow(n int) {
	for i := 0; i < n; i++ {
		w.Grid = append(w.Grid, Row{})
	}
}

// LastRow returns the last row of the window.
func (w *Window) LastRow() Row {
	return w.Grid[len(w.Grid)-1]
}

// Render renders the window.
func (w *Window) Render() string {
	if w.IsAltScreen {
		return w.RenderAltScreen()
	}

	var buf strings.Builder
	// render title
	{
		title := w.Title.Render()
		title = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(2).Render(title)
		buf.WriteString(title)
		buf.WriteString("\n")
	}

	// render grid
	{
		for _, r := range w.Grid {
			buf.WriteString(r.Render())
			buf.WriteString("\n")
		}
	}

	// render footer
	{
		footer := w.Footer.Render()
		buf.WriteString(footer)
	}

	return w.Style.Render(buf.String())
}

func (w *Window) RenderAltScreen() string {
	var buf strings.Builder
	// render title
	{
		style := lipgloss.NewStyle().MarginTop(w.Height / 10)
		title := style.Render(w.Title.Render())
		buf.WriteString(title)
		buf.WriteString("\n")
	}

	// render grid
	{
		style := lipgloss.NewStyle()

		for i, r := range w.Grid {
			var s string
			if i == 0 {
				s = style.Copy().MarginTop(int(float64(w.Height) * 0.18)).Render(r.Render())
			} else {
				s = style.Copy().Render(r.Render())
			}
			buf.WriteString(s)
			buf.WriteString("\n")
		}
	}

	// render footer
	{
		style := lipgloss.NewStyle().MarginTop(w.Height - lipgloss.Height(buf.String()) - 1)
		footer := style.Render(w.Footer.Render())
		buf.WriteString(footer)
	}

	return w.Style.Render(buf.String())
}

// Row indicates a row in the window, it contains multiple blocks.
type Row []Blocker

func (r Row) Width() int {
	width := 0
	for _, b := range r {
		width += b.Width()
	}
	return width
}

func (r Row) Height() int {
	height := 0
	for _, b := range r {
		if height < b.Height() {
			height = b.Height()
		}
	}
	return height
}

func (r Row) Render() string {
	style := lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(1)

	var bs []string
	for _, b := range r {
		bs = append(bs, style.Render(b.Render()))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, bs...)
}
