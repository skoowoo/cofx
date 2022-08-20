package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cofunclabs/cofunc/pkg/feedbackid"
	"github.com/cofunclabs/cofunc/service"
	"github.com/cofunclabs/cofunc/service/exported"
)

var (
	runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	doneMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func startRunningUI(svc *service.SVC, fi *exported.FlowInsight) error {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	model := runningModel{
		svc:     svc,
		fi:      fi,
		spinner: s,
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		),
	}

	rand.Seed(time.Now().Unix())
	return tea.NewProgram(model).Start()
}

type getFlowInsightErr error

func getFlowInsightCmd(svc *service.SVC, id string) tea.Cmd {
	d := time.Millisecond * time.Duration(rand.Intn(500))
	return tea.Tick(d, func(t time.Time) tea.Msg {
		fi, err := svc.InsightFlow(context.Background(), feedbackid.WrapID(id))
		if err != nil {
			return getFlowInsightErr(err)
		}
		return &fi
	})
}

type runningModel struct {
	svc      *service.SVC
	fi       *exported.FlowInsight
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

func (m runningModel) Init() tea.Cmd {
	return tea.Batch(getFlowInsightCmd(m.svc, m.fi.ID), m.spinner.Tick)
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
			getFlowInsightCmd(m.svc, m.fi.ID),
		)
	case getFlowInsightErr:
		return m, getFlowInsightCmd(m.svc, m.fi.ID)
	}
	return m, nil
}

func (m runningModel) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\nExecute the flow %s\n\n", m.fi.ID))

	for _, n := range m.fi.Nodes {
		if n.Status == "RUNNING" {
			builder.WriteString(fmt.Sprintf("%s #%d %s\n", m.spinner.View(), n.Step, runningStyle.Render(n.Name)))
		} else if n.Status == "STOPPED" {
			builder.WriteString(fmt.Sprintf("%s #%d %s\n", doneMark.String(), n.Step, n.Name))
		} else {
			builder.WriteString(fmt.Sprintf("%s #%d %s\n", " ", n.Step, n.Name))
		}
	}

	if m.done {
		builder.WriteString(doneStyle.Render(fmt.Sprintf("\nDone! Executed %d functions.\n", m.fi.Total)))
	}

	// spin := m.spinner.View() + " "
	// prog := m.progress.View()

	return builder.String()
}
