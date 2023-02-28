package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/skoowoo/cofx/pkg/nameid"
	pretty "github.com/skoowoo/cofx/pkg/pretty"
	"github.com/skoowoo/cofx/service"
	"github.com/skoowoo/cofx/service/exported"
)

func prunEntry(nameorid nameid.NameOrID) error {
	svc := service.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fid nameid.ID

	path, fid, err := svc.LookupFlowl(ctx, nameorid)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := svc.AddFlow(ctx, fid, f); err != nil {
		return err
	}
	if _, err := svc.ReadyFlow(ctx, fid, nil); err != nil {
		return err
	}

	var (
		lasterr error
		wg      sync.WaitGroup
	)
	wg.Add(2)
	// start the ui in a goroutine
	go func() {
		defer func() {
			wg.Done()
			cancel()
			// cancel() be used to stop the event trigger goroutine
			svc.CancelRunningFlow(ctx, fid)
		}()

		if err := startPrunView(func() (*exported.FlowRunningInsight, error) {
			fi, err := svc.InsightFlow(ctx, fid)
			return &fi, err
		}); err != nil {
			lasterr = err
		}
	}()

	time.Sleep(time.Second)
	// start the flow in a goroutine
	go func() {
		defer wg.Done()
		err := svc.StartFlowOrEventFlow(ctx, fid)
		if err != nil {
			lasterr = err
		}
		prunCmdExited = true
	}()
	wg.Wait()

	if lasterr != nil {
		os.Exit(-1)
	}

	return nil
}

var prunCmdExited bool

func startPrunView(get func() (*exported.FlowRunningInsight, error)) error {
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
	}

	rand.Seed(time.Now().Unix())
	return tea.NewProgram(model).Start()
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

		if prunCmdExited && !m.fullscreen {
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

func (m prunModel) View() string {
	window := pretty.NewWindow(m.height, m.width, false)
	window.SetTitle(pretty.NewTitleBlock("Pretty Run Flow: "+m.fi.Name, m.fi.ID))

	headers := []string{pretty.IconSpace.String(), "STEP", "SEQ", "NAME", "DRIVER", "RUNS", "DURATION"}
	var values [][]string
	for _, n := range m.fi.Nodes {
		if n.Status == "RUNNING" {
			values = append(values, []string{
				m.spinner.View(),
				strconv.Itoa(n.Step),
				strconv.Itoa(n.Seq),
				n.Name + " ➜ " + n.Function,
				n.Driver,
				strconv.Itoa(n.Runs),
				fmt.Sprintf("%dms", n.Duration),
			})
		} else if n.Status == "STOPPED" {
			icon := pretty.IconOK
			if n.LastError != nil {
				icon = pretty.IconFailed
			}
			values = append(values, []string{
				icon.String(),
				strconv.Itoa(n.Step),
				strconv.Itoa(n.Seq),
				n.Name + " ➜ " + n.Function,
				n.Driver,
				strconv.Itoa(n.Runs),
				fmt.Sprintf("%dms", n.Duration),
			})
		} else {
			values = append(values, []string{
				pretty.IconSpace.String(),
				strconv.Itoa(n.Step),
				strconv.Itoa(n.Seq),
				n.Name + " ➜ " + n.Function,
				n.Driver,
				strconv.Itoa(n.Runs),
				fmt.Sprintf("%dms", n.Duration),
			})
		}
	}
	window.AppendBlock(pretty.NewTableBlock(headers, values))
	window.AppendNewRow(1)

	if m.done {
		s := "\n" + pretty.IconOK.String() + fmt.Sprintf("Done! Duration: %dms", m.fi.Duration)
		window.AppendBlock(pretty.NewTextBlock(s))
	} else {
		s := "\n" + m.spinner.View() + fmt.Sprintf(" Running... Duration: %dms", m.fi.Duration)
		window.AppendBlock(pretty.NewTextBlock(s))
	}

	return window.Render()
}
