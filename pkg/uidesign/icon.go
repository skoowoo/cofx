package uidesign

import "github.com/charmbracelet/lipgloss"

var (
	IconRight = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Width(2).SetString("➜")
	IconCycle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Width(2).SetString("●")
)
