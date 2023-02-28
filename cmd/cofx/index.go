package main

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"context"
	"strconv"

	"github.com/skoowoo/cofx/config"
	pretty "github.com/skoowoo/cofx/pkg/pretty"
	"github.com/skoowoo/cofx/service"

	tea "github.com/charmbracelet/bubbletea"
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
	window := pretty.NewWindow(m.height, m.width, true)
	window.SetTitle(pretty.NewTitleBlock("CoFx", pretty.ShadeText("Turn boring stuff into low code ...", 0), true))

	version := "v0.0.2"
	help := version + " https://github.com/skoowoo/cofx • Press any key to exit"
	window.SetFooter(pretty.NewFooterBlock(help))

	kvs := [][]string{
		{"Standard Functions", strconv.Itoa(m.functions)},
		{"Available Flows", strconv.Itoa(m.flows)},
		{"Home Directory", m.homeDir},
	}
	window.AppendBlock(pretty.NewKvsBlock(kvs...))

	return window.Render()
}
