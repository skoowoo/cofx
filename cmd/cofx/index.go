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
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
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
		Foreground(lipgloss.Color("15"))
	name := "cofx"
	name = nameStyle.Render(name)

	// slogan
	sloganStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)
	slogan := "Turn boring stuff into low code ..."
	slogan = sloganStyle.Render(uidesign.ShadeText(slogan, 2))

	// metrics
	functions := lipgloss.JoinVertical(lipgloss.Left, strconv.Itoa(m.functions), "Standard Functions")
	flows := lipgloss.JoinVertical(lipgloss.Left, strconv.Itoa(m.flows), "Available Flows")
	homeDir := lipgloss.JoinVertical(lipgloss.Left, m.homeDir, "CoFx Home Directory")

	maxWidth := max(lipgloss.Width(functions), lipgloss.Width(flows), lipgloss.Width(homeDir))
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2).
		Height(4).
		Width(maxWidth + 5)

	grid := uidesign.ColorGrid(14, 8)
	c1 := grid[4][0]
	c2 := grid[4][1]
	c3 := grid[4][2]
	sf := blockStyle.Copy().Background(lipgloss.Color(c1)).Render(functions)
	af := blockStyle.Copy().Background(lipgloss.Color(c2)).Render(flows)
	hd := blockStyle.Copy().Background(lipgloss.Color(c3)).Render(homeDir)
	metrics := lipgloss.JoinHorizontal(lipgloss.Bottom, af, sf, hd)
	metrics = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(metrics)

	// version
	version := "v0.0.1"
	from := "https://github.com/cofxlabs/cofx"
	version = version + " " + from
	version = lipgloss.NewStyle().MarginLeft(m.width/2 - len(version)/2).MarginTop(m.height / 10 * 5).Render(version)
	version = uidesign.ColorGrey.Render(version)

	return name + "\n\n" + slogan + "\n\n\n\n" + metrics + "\n" + version
}
