package main

import "github.com/charmbracelet/lipgloss"

var (
	docStyle         = lipgloss.NewStyle().MarginTop(1).MarginBottom(0).MarginLeft(2).MarginRight(0)
	flowNameStyle    = lipgloss.NewStyle().Width(20)
	flowIDStyle      = lipgloss.NewStyle().Width(35)
	stepStyle        = lipgloss.NewStyle().Width(4)
	seqStyle         = lipgloss.NewStyle().Width(4)
	runsStyle        = lipgloss.NewStyle().Width(6)
	driverStyle      = lipgloss.NewStyle().Width(8)
	runningNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	nameStyle        = lipgloss.NewStyle()

	doneMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	errorMark = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("✗")
)

type viewErrorMessage error
