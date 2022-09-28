package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cofxlabs/cofx/config"
	"github.com/cofxlabs/cofx/service"
)

func listFlows() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables := svc.ListAvailables(ctx)
	sort.Slice(availables, func(i, j int) bool { return availables[i].Name < availables[j].Name })

	// calculate the max length of flow's source field
	var max int = 20
	for i, f := range availables {
		source := strings.TrimPrefix(f.Source, config.PrivateFlowlDir())
		if source == f.Source {
			source = strings.TrimPrefix(f.Source, config.BaseFlowlDir())
		}
		availables[i].Source = source
		if max < len(source) {
			max = len(source)
		}
	}
	sourceStyle := lipgloss.NewStyle().Width(max + 2)

	// here is title
	fmt.Fprintln(os.Stdout, "\n"+
		colorGrey.Render(iconSpace.String()+
			flowNameStyle.Render("FLOW NAME")+
			flowIDStyle.Render("FLOW ID")+
			sourceStyle.Render("SOURCE")+
			"DESC"))

	for _, f := range availables {
		var s string
		if f.Total == -1 {
			s = iconCircleFailed.String() +
				flowNameStyle.Foreground(lipgloss.Color("222")).Render(f.Name) +
				flowIDStyle.Render(f.ID) +
				sourceStyle.Render(f.Source) +
				colorRed.MaxWidth(30).Render(f.Desc)
		} else {
			s = iconCircleOk.String() +
				flowNameStyle.Foreground(lipgloss.Color("222")).Render(f.Name) +
				flowIDStyle.Render(f.ID) +
				sourceStyle.Render(f.Source) +
				lipgloss.NewStyle().MaxWidth(60).Render(f.Desc)
		}
		fmt.Fprintln(os.Stdout, s)
	}
	fmt.Fprintf(os.Stdout, "\n")
	return nil
}
