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
	iconStyle  = lipgloss.NewStyle().Width(2)
	iconSpace  = lipgloss.NewStyle().Width(2).SetString(" ")
	iconOK     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓").Width(2)
	iconFailed = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("✗").Width(2)
	iconCircle = lipgloss.NewStyle().Foreground(lipgloss.Color("25")).SetString("●").Width(2)
	colorGrey  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
