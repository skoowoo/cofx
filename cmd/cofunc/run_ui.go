package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cofunclabs/cofunc/service/exported"
)

var (
	runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	doneMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func startRunningUI(get func() (*exported.FlowInsight, error)) error {
	fi, err := get()
	if err != nil {
		return err
	}

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	model := runningModel{
		spinner: s,
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		),
		fi: fi,
		getCmd: tea.Tick(time.Millisecond*time.Duration(rand.Intn(500)), func(t time.Time) tea.Msg {
			fi, err := get()
			if err != nil {
				return getFlowInsightErr(err)
			}
			return fi
		}),
	}

	rand.Seed(time.Now().Unix())
	return tea.NewProgram(model).Start()
}

type getFlowInsightErr error

type runningModel struct {
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
	fi       *exported.FlowInsight
	getCmd   tea.Cmd
}

func (m runningModel) Init() tea.Cmd {
	return tea.Batch(m.getCmd, m.spinner.Tick)
}

func (m runningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	case *exported.FlowInsight:
		m.fi = msg
		if m.fi.Done == m.fi.Total {
			m.done = true
			return m, tea.Quit
		}

		// Update progress bar
		progressCmd := m.progress.SetPercent(float64(m.fi.Done) / float64(m.fi.Total))

		return m, tea.Batch(
			progressCmd,
			m.getCmd,
		)
	case getFlowInsightErr:
		return m, m.getCmd
	}
	return m, nil
}

func (m runningModel) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\nExecute the flow: %s\n\n", m.fi.ID))

	for _, n := range m.fi.Nodes {
		if n.Status == "RUNNING" {
			builder.WriteString(fmt.Sprintf("%s #%d %s (%d)\n", m.spinner.View(), n.Step, runningStyle.Render(n.Name), n.Runs))
		} else if n.Status == "STOPPED" {
			builder.WriteString(fmt.Sprintf("%s #%d %s (%d)\n", doneMark.String(), n.Step, n.Name, n.Runs))
		} else {
			builder.WriteString(fmt.Sprintf("%s #%d %s (%d)\n", " ", n.Step, n.Name, n.Runs))
		}
	}

	if m.done {
		builder.WriteString(doneStyle.Render("\nDone!\n"))
	} else {
		builder.WriteString(doneStyle.Render(m.spinner.View() + " Running...\n"))
	}

	// spin := m.spinner.View() + " "
	// prog := m.progress.View()

	return builder.String()
}
