package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	co "github.com/cofxlabs/cofx"
	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/pkg/output"
	"github.com/cofxlabs/cofx/runtime"
	"github.com/cofxlabs/cofx/service"
	"github.com/cofxlabs/cofx/service/exported"
)

func runEntry(nameorid nameid.NameOrID) error {
	svc := service.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fid nameid.ID

	// If the argument 'nameorid' not contains the suffix ".flowl", We will treat it as a flow name or id, so we will lookup the flowl source path through
	// name or id.
	// if the argument 'nameorid' contains the suffix ".flowl", we will treat it as a full path of the flowl file, so can open it directly.
	fp := nameorid.String()
	if !co.IsFlowl(fp) {
		id, err := svc.LookupID(ctx, nameorid)
		if err != nil {
			return err
		}
		meta, err := svc.GetAvailableMeta(ctx, id)
		if err != nil {
			return err
		}
		fp = meta.Source
		fid = id
	} else {
		fid = nameid.New(co.FlowlPath2Name(fp))
	}
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	if err := svc.AddFlow(ctx, fid, f); err != nil {
		return err
	}

	lineC := make(chan string, 100)
	out := &output.Output{
		W: nil,
		HandleFunc: func(line []byte) {
			line = bytes.TrimSuffix(line, []byte{'\n'})
			lineC <- string(line)
		},
	}
	if _, err := svc.ReadyFlow(ctx, fid, out); err != nil {
		return err
	}

	var errs []error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer func() {
			// Invoke the defer statement to make sure the ui goroutine exited, when the flow exited.
			close(lineC)
			wg.Done()
		}()

		if err := svc.StartFlowOrEventFlow(ctx, fid); err != nil {
			errs = append(errs, err)
			return
		}
	}()

	// start the 'run' ui
	go func() {
		defer func() {
			// When ui exited, need to cancel the context to make sure the whole flow exit.
			svc.CancelRunningFlow(ctx, fid)
			wg.Done()
		}()
		rand.Seed(time.Now().UnixNano())

		m := newRunModel()
		m.getMsgFunc = func() tea.Cmd {
			return tea.Tick(time.Millisecond*time.Duration(rand.Intn(50)), func(t time.Time) tea.Msg {
				insight, err := svc.InsightFlow(ctx, fid)
				if err != nil {
					return err
				}
				// insight.JsonWrite(os.Stdout)
				return runGetMsg{
					status: insight,
				}
			})
		}
		m.subMsgFunc = func() tea.Cmd {
			return func() tea.Msg {
				l, ok := <-lineC
				if !ok {
					return runSubMsg{
						exit: true,
					}
				}
				return runSubMsg{
					l: l,
				}
			}
		}

		if err := tea.NewProgram(m).Start(); err != nil {
			errs = append(errs, err)
			return
		}
	}()
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

type runSubMsg struct {
	l    string
	exit bool
}

type runGetMsg struct {
	status exported.FlowRunningInsight
}

type runModel struct {
	width      int
	height     int
	spinner    spinner.Model
	progress   progress.Model
	getMsgFunc func() tea.Cmd
	subMsgFunc func() tea.Cmd
	totalNum   int
	doneNum    int
	nodes      []exported.NodeRunningInsight
}

func newRunModel() runModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(80),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return runModel{
		spinner:  s,
		progress: p,
	}
}

func (m runModel) Init() tea.Cmd {
	return tea.Batch(m.getMsgFunc(), m.subMsgFunc(), m.spinner.Tick)
}

func (m runModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case runSubMsg:
		if msg.exit {
			return m, tea.Quit
		}
		return m, tea.Batch(m.subMsgFunc(), tea.Printf("%s", msg.l))
	case runGetMsg:
		m.doneNum = msg.status.Done
		m.totalNum = msg.status.Total
		m.nodes = msg.status.Nodes
		// Update progress bar
		progressCmd := m.progress.SetPercent(float64(m.doneNum) / float64(m.totalNum))

		return m, tea.Batch(progressCmd, m.getMsgFunc())
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
	case error:
		tea.Println(msg)
		return m, tea.Quit
	}
	return m, nil
}

var executing = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))

func (m runModel) View() string {
	w := lipgloss.Width(fmt.Sprintf("%d", m.totalNum))

	doneCount := fmt.Sprintf(" %*d/%*d", w, m.doneNum, w, m.totalNum)
	spin := m.spinner.View() + " "

	var names []string
	for _, n := range m.nodes {
		if n.Status == string(runtime.StatusRunning) {
			names = append(names, n.Name)
		}
	}
	running := fmt.Sprintf("Running %s", executing.Render(strings.Join(names, ", ")))

	if m.width/3 < 80 {
		m.progress.Width = m.width / 3
	}
	prog := m.progress.View()

	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+doneCount+running))

	gap := strings.Repeat(" ", cellsAvail)

	return spin + running + gap + prog + doneCount
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
