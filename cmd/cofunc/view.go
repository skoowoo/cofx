package main

import "github.com/charmbracelet/lipgloss"

type viewErrorMessage error

var (
	docStyle = lipgloss.NewStyle().MarginTop(1).MarginBottom(0).MarginLeft(2).MarginRight(0)

	// flow
	flowNameStyle = lipgloss.NewStyle().Width(20)
	flowIDStyle   = lipgloss.NewStyle().Width(35)

	// node & function
	stepStyle        = lipgloss.NewStyle().Width(5)
	seqStyle         = lipgloss.NewStyle().Width(4)
	runsStyle        = lipgloss.NewStyle().Width(6)
	driverStyle      = lipgloss.NewStyle().Width(8)
	nameStyle        = lipgloss.NewStyle().Width(20)
	runningNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))

	// common
	colorGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	colorGrey  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	colorRed   = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))

	iconStyle  = lipgloss.NewStyle().Width(2)
	iconSpace  = lipgloss.NewStyle().Width(2).SetString(" ")
	iconOK     = colorGreen.Copy().Width(2).SetString("✓")
	iconFailed = colorRed.Copy().Width(2).SetString("✗")
)
