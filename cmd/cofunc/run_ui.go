package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cofunclabs/cofunc/service/exported"
)

var runCmdExited bool

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
		getCmd: tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
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
		}

		// Update progress bar
		progressCmd := m.progress.SetPercent(float64(m.fi.Done) / float64(m.fi.Total))

		if runCmdExited {
			return m, tea.Quit
		}

		return m, tea.Batch(
			progressCmd,
			m.getCmd,
		)
	case getFlowInsightErr:
		return m, m.getCmd
	}
	return m, nil
}

var (
	doneStyle        = lipgloss.NewStyle().MarginLeft(0).MarginTop(2)
	doneMark         = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	errorMark        = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).SetString("✗")
	stepStyle        = lipgloss.NewStyle().Width(4)
	seqStyle         = lipgloss.NewStyle().Width(4)
	runsStyle        = lipgloss.NewStyle().Width(6)
	runningNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	nameStyle        = lipgloss.NewStyle()
	driverStyle      = lipgloss.NewStyle().Width(8)
)

func maxNameWidth(nodes []exported.NodeInsight) int {
	max := 20
	for _, n := range nodes {
		w := lipgloss.Width(n.Name + " ➜ " + n.Function)
		if w > max {
			max = w + 2
		}
	}
	return max
}

func (m runningModel) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\nFLOW NAME: %s", m.fi.Name))
	builder.WriteString(fmt.Sprintf("\nFLOW ID:   %s\n\n", m.fi.ID))

	nameMaxWidth := maxNameWidth(m.fi.Nodes)

	builder.WriteString(
		fmt.Sprintf("%s %s  %s %s %s %s DURATION\n",
			" ",
			stepStyle.Render("STEP"),
			seqStyle.Render("SEQ"),
			nameStyle.Width(nameMaxWidth).Render("NAME"),
			driverStyle.Render("DRIVER"),
			runsStyle.Render("RUNS"),
		))

	for _, n := range m.fi.Nodes {
		if n.Status == "RUNNING" {
			builder.WriteString(
				fmt.Sprintf("%s #%s %s %s %s %s %dms\n",
					m.spinner.View(),
					stepStyle.Render(strconv.Itoa(n.Step)),
					seqStyle.Render(strconv.Itoa(n.Seq)),
					runningNameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function),
					driverStyle.Render(n.Driver),
					runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)),
					n.Duration))
		} else if n.Status == "STOPPED" {
			mark := doneMark
			if n.LastError != nil {
				mark = errorMark
			}
			builder.WriteString(
				fmt.Sprintf("%s #%s %s %s %s %s %dms\n",
					mark.String(),
					stepStyle.Render(strconv.Itoa(n.Step)),
					seqStyle.Render(strconv.Itoa(n.Seq)),
					nameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function),
					driverStyle.Render(n.Driver),
					runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)),
					n.Duration))
		} else {
			builder.WriteString(
				fmt.Sprintf("%s #%s %s %s %s %s %dms\n",
					" ",
					stepStyle.Render(strconv.Itoa(n.Step)),
					seqStyle.Render(strconv.Itoa(n.Seq)),
					nameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function),
					driverStyle.Render(n.Driver),
					runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)),
					n.Duration))
		}
	}

	if m.done {
		builder.WriteString(doneStyle.Render(doneMark.String() + " Done!" + fmt.Sprintf(" Duration: %dms\n\n", m.fi.Duration)))
	} else {
		builder.WriteString(doneStyle.Render(m.spinner.View() + " Running..." + fmt.Sprintf(" Duration: %dms\n\n", m.fi.Duration)))
	}

	// spin := m.spinner.View() + " "
	// prog := m.progress.View()

	return builder.String()
}
