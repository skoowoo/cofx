package uidesign

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	ColorGrey = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
)

func ColorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}

func ShadeText(text string, yoffset ...int) string {
	var style = lipgloss.NewStyle()
	var buf strings.Builder
	grid := ColorGrid(14, 8)
	for i, ch := range text {
		x := i % 14
		y := i / 14
		if len(yoffset) > 0 {
			off := yoffset[0] % 8
			y += off
		}
		y %= 8

		var j int
		if y%2 == 0 {
			j = x
		} else {
			j = 13 - x
		}
		color := grid[y][j]
		buf.WriteString(style.Foreground(lipgloss.Color(color)).Render(string(ch)))
	}
	return buf.String()
}
