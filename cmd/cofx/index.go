package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cofxlabs/cofx/config"
	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/service"
	"github.com/cofxlabs/cofx/service/exported"
)

func indexEntry() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables := svc.ListAvailables(ctx)

	//  execute 'cofx' command without any args or sub-command
	for {
		selected, err := startListingView(availables)
		if err != nil {
			return err
		}

		// to run the selected flow
		if selected.Source != "" {
			err := prunEntry(nameid.NameOrID(selected.Source), true)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}

func startListingView(flows []exported.FlowMetaInsight) (exported.FlowMetaInsight, error) {
	items := make([]list.Item, 0, len(flows))
	for _, f := range flows {
		items = append(items, flowItem(f))
	}

	keys := newAdditionalKeyMap()
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "All Available Flows"
	l.Styles.Title = list.DefaultStyles().Title
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.toggleQuit,
			keys.toggleRun,
			keys.toggleEdit,
		}
	}

	model := indexModel{
		list: l,
		keys: keys,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	ret, err := p.StartReturningModel()
	if err != nil {
		return exported.FlowMetaInsight{}, err
	}

	if m, ok := ret.(indexModel); ok && m.flowToRun.Name != "" {
		return exported.FlowMetaInsight(m.flowToRun), nil
	}
	return exported.FlowMetaInsight{}, nil
}

type indexModel struct {
	list      list.Model
	keys      keyMap
	flowToRun flowItem
}

func (m indexModel) Init() tea.Cmd {
	return nil
}

func (m indexModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.toggleQuit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.toggleRun) {
			m.flowToRun = m.getSelected()
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.toggleEdit) {
			return m, openEditor(m.getSelected().Source)
		}
	case editorFinishedMsg:
		// parse this flow again
		svc := service.New()
		insight, err := svc.GetAvailableMeta(context.Background(), nameid.New(m.getSelected().Name))
		if err != nil {
			m.updateSelected(func(selected *flowItem) { selected.Desc = err.Error() })
		} else {
			m.replaceSelected(flowItem(insight))
		}
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m indexModel) View() string {
	return docStyle.Render(m.list.View())
}

type flowItem exported.FlowMetaInsight

func (f flowItem) Title() string {
	return fmt.Sprintf("%s %s", f.Name, f.ID)
}

func (f flowItem) Description() string {
	return fmt.Sprintf("%d %s %s", f.Total, strings.TrimPrefix(f.Source, config.FlowSourceDir()), f.Desc)
}

func (f flowItem) FilterValue() string {
	return f.Name
}

func (m indexModel) getSelected() flowItem {
	return m.list.SelectedItem().(flowItem)
}

func (m indexModel) updateSelected(f func(selected *flowItem)) {
	selected := m.getSelected()
	f(&selected)
	m.list.SetItem(m.list.Index(), selected)
}

func (m indexModel) replaceSelected(item flowItem) {
	m.list.SetItem(m.list.Index(), item)
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

type keyMap struct {
	toggleQuit key.Binding
	toggleRun  key.Binding
	toggleEdit key.Binding
}

func newAdditionalKeyMap() keyMap {
	return keyMap{
		toggleQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		toggleRun: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "run flow"),
		),
		toggleEdit: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "edit flow"),
		),
	}
}
