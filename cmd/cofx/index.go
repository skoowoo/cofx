package main

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"context"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cofxlabs/cofx/config"
	"github.com/cofxlabs/cofx/pkg/uidesign"
	"github.com/cofxlabs/cofx/service"
)

func indexEntry() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	functions := svc.ListStdFunctions(ctx)
	flows := svc.ListAvailables(ctx)
	m := indexModel{
		functions: len(functions),
		flows:     len(flows),
		homeDir:   config.HomeDir(),
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Start()
}

type indexModel struct {
	height    int
	width     int
	flows     int
	functions int
	homeDir   string
}

func (m indexModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m indexModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, nil
}

func (m indexModel) View() string {
	// name
	nameStyle := lipgloss.NewStyle().
		Width(m.width).
		Bold(true).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(m.height / 10).
		Foreground(lipgloss.Color("15"))
	name := "cofx"
	name = nameStyle.Render(name)

	// slogan
	sloganStyle := lipgloss.NewStyle().
		MarginTop(2).
		Width(m.width).
		Align(lipgloss.Center)
	slogan := "Turn boring stuff into low code ..."
	slogan = sloganStyle.Render(uidesign.ShadeText(slogan, 0))

	// metrics
	functions := lipgloss.JoinVertical(lipgloss.Left, strconv.Itoa(m.functions), uidesign.ColorGrey1.Render("Standard Functions"))
	flows := lipgloss.JoinVertical(lipgloss.Left, strconv.Itoa(m.flows), uidesign.ColorGrey1.Render("Available Flows"))
	homeDir := lipgloss.JoinVertical(lipgloss.Left, m.homeDir, uidesign.ColorGrey1.Render("CoFx Home Directory"))

	maxWidth := max(lipgloss.Width(functions), lipgloss.Width(flows), lipgloss.Width(homeDir))
	blockStyle := lipgloss.NewStyle().
		BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("99")).
		Align(lipgloss.Left).
		MarginLeft(0).MarginRight(1).
		Width(maxWidth)
	blockLayoutStyle := lipgloss.NewStyle().Margin(int(float64(m.height)*0.18), 1, 1, 1)
	af := blockStyle.Copy().Render(flows)
	af = blockLayoutStyle.Render(af)

	sf := blockStyle.Copy().Render(functions)
	sf = blockLayoutStyle.Render(sf)

	hd := blockStyle.Copy().Render(homeDir)
	hd = blockLayoutStyle.Render(hd)

	metrics := lipgloss.JoinHorizontal(lipgloss.Bottom, af, sf, hd)
	metrics = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(metrics)

	// version
	remainH := m.height - lipgloss.Height(name) - lipgloss.Height(slogan) - lipgloss.Height(metrics)

	versionStyle := lipgloss.NewStyle().
		MarginTop(remainH + 1).
		Width(m.width).
		Align(lipgloss.Center)
	version := "v0.0.1"
	from := "https://github.com/cofxlabs/cofx"
	help := "Press any key to exit"
	version = version + " " + from + " • " + help
	version = versionStyle.Render(uidesign.ColorGrey3.Render(version))

	return name + slogan + metrics + version
}
