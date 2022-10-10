package main

import "github.com/charmbracelet/lipgloss"

type viewErrorMessage error

var (
	// function
	funcNameStyle = lipgloss.NewStyle().Width(30)

	// flow
	flowNameStyle = lipgloss.NewStyle().Width(20)
	flowIDStyle   = lipgloss.NewStyle().Width(35)

	// common
	colorGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	colorGrey  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	colorRed   = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))

	iconSpace        = lipgloss.NewStyle().Width(2).SetString(" ")
	iconCircleOk     = colorGreen.Copy().Width(2).SetString("●")
	iconCircleFailed = colorRed.Copy().Width(2).SetString("●")
)
