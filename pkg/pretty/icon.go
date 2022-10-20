package pretty

import "github.com/charmbracelet/lipgloss"

var (
	IconSpace           = lipgloss.NewStyle().Width(2).SetString(" ")
	IconRight           = ColorGreen.Copy().Width(2).SetString("➜")
	IconCycle           = ColorGreen.Copy().Width(2).SetString("●")
	IconOK              = ColorGreen.Copy().Width(2).SetString("✓")
	IconFailed          = ColorRed.Copy().Width(2).SetString("✗")
	IconMinCircleOk     = ColorGreen.Copy().Width(2).SetString("·")
	IconMinCircleFailed = ColorRed.Copy().Width(2).SetString("·")
)
