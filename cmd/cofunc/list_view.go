package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cofunclabs/cofunc/service/exported"
)

func startListingView(flows []exported.FlowMetaInsight) (exported.FlowMetaInsight, error) {
	items := []list.Item{}
	for _, f := range flows {
		items = append(items, flowItem(f))
	}

	model := listFlowModel{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	model.list.Title = "All Available Flows"

	p := tea.NewProgram(model, tea.WithAltScreen())
	ret, err := p.StartReturningModel()
	if err != nil {
		return exported.FlowMetaInsight{}, err
	}

	if m, ok := ret.(listFlowModel); ok && m.selected.Name != "" {
		return m.selected, nil
	}
	return exported.FlowMetaInsight{}, nil
}

type listFlowModel struct {
	list     list.Model
	selected exported.FlowMetaInsight
}

func (m listFlowModel) Init() tea.Cmd {
	return nil
}

func (m listFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+r" {
			m.selected = exported.FlowMetaInsight(m.list.SelectedItem().(flowItem))
			return m, tea.Quit
		}
		if msg.String() == "ctrl+e" {
			return m, openEditor(m.list.SelectedItem().(flowItem).Source)
		}
	case editorFinishedMsg:

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listFlowModel) View() string {
	return docStyle.Render(m.list.View())
}

type flowItem exported.FlowMetaInsight

func (f flowItem) Title() string {
	return fmt.Sprintf("%s %s", f.Name, f.ID)
}

func (f flowItem) Description() string {
	return fmt.Sprintf("%d %s %s", f.Total, f.Source, f.Desc)
}

func (f flowItem) FilterValue() string {
	return f.Name
}

type editorFinishedMsg struct {
	err error
}

func openEditor(file string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, file) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
