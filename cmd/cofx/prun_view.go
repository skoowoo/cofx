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
	"github.com/cofxlabs/cofx/service/exported"
)

var runCmdExited bool

func startPrunView(fullscreen bool, get func() (*exported.FlowRunningInsight, error)) error {
	fi, err := get()
	if err != nil {
		return err
	}

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	model := prunModel{
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
				return viewErrorMessage(err)
			}
			return fi
		}),
		fullscreen: fullscreen,
	}

	rand.Seed(time.Now().Unix())
	if fullscreen {
		return tea.NewProgram(model, tea.WithAltScreen()).Start()
	} else {
		return tea.NewProgram(model).Start()
	}
}

type prunModel struct {
	width      int
	height     int
	spinner    spinner.Model
	progress   progress.Model
	done       bool
	fi         *exported.FlowRunningInsight
	getCmd     tea.Cmd
	fullscreen bool
}

func (m prunModel) Init() tea.Cmd {
	return tea.Batch(m.getCmd, m.spinner.Tick)
}

func (m prunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case *exported.FlowRunningInsight:
		m.fi = msg
		if m.fi.Done == m.fi.Total {
			m.done = true
		} else {
			m.done = false
		}

		// Update progress bar
		progressCmd := m.progress.SetPercent(float64(m.fi.Done) / float64(m.fi.Total))

		if runCmdExited && !m.fullscreen {
			return m, tea.Quit
		}

		return m, tea.Batch(
			progressCmd,
			m.getCmd,
		)
	case viewErrorMessage:
		return m, m.getCmd
	}
	return m, nil
}

func maxNameWidth(nodes []exported.NodeRunningInsight) int {
	max := 20
	for _, n := range nodes {
		w := lipgloss.Width(n.Name + " ➜ " + n.Function)
		if w > max {
			max = w + 2
		}
	}
	return max
}

func (m prunModel) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  FLOW NAME: %s\n", m.fi.Name))
	builder.WriteString(fmt.Sprintf("  FLOW ID:   %s\n\n", m.fi.ID))

	nameMaxWidth := maxNameWidth(m.fi.Nodes)

	builder.WriteString(colorGrey.Render(iconSpace.String() +
		stepStyle.Render("STEP") +
		seqStyle.Render("SEQ") +
		nameStyle.Width(nameMaxWidth).Render("NAME") +
		driverStyle.Render("DRIVER") +
		runsStyle.Render("RUNS")))

	builder.WriteString("\n")

	for _, n := range m.fi.Nodes {
		if n.Status == "RUNNING" {
			builder.WriteString(iconStyle.Render(m.spinner.View()) +
				stepStyle.Render(strconv.Itoa(n.Step)) +
				seqStyle.Render(strconv.Itoa(n.Seq)) +
				runningNameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function) +
				driverStyle.Render(n.Driver) +
				runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)) +
				fmt.Sprintf("%dms", n.Duration) +
				"\n")
		} else if n.Status == "STOPPED" {
			mark := iconOK
			if n.LastError != nil {
				mark = iconFailed
			}
			builder.WriteString(mark.String() +
				stepStyle.Render(strconv.Itoa(n.Step)) +
				seqStyle.Render(strconv.Itoa(n.Seq)) +
				nameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function) +
				driverStyle.Render(n.Driver) +
				runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)) +
				fmt.Sprintf("%dms", n.Duration) +
				"\n")
		} else {
			builder.WriteString(iconSpace.String() +
				stepStyle.Render(strconv.Itoa(n.Step)) +
				seqStyle.Render(strconv.Itoa(n.Seq)) +
				nameStyle.Width(nameMaxWidth).Render(n.Name+" ➜ "+n.Function) +
				driverStyle.Render(n.Driver) +
				runsStyle.Render(fmt.Sprintf("(%d)", n.Runs)) +
				fmt.Sprintf("%dms", n.Duration) +
				"\n")
		}
	}

	if m.done {
		builder.WriteString("\n\n" + iconOK.String() + fmt.Sprintf(" Done! Duration: %dms\n\n", m.fi.Duration))
	} else {
		builder.WriteString("\n\n" + m.spinner.View() + fmt.Sprintf(" Running... Duration: %dms\n\n", m.fi.Duration))
	}

	// spin := m.spinner.View() + " "
	// prog := m.progress.View()

	return docStyle.Render(builder.String())
}
