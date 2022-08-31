package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/cofunclabs/cofunc/service"
)

func listFlows() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables := svc.ListAvailables(ctx)

	// calculate the max length of flow's source field
	var max int = 20
	for _, f := range availables {
		if max < len(f.Source) {
			max = len(f.Source)
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
			s = iconFailed.String() +
				flowNameStyle.Render(f.Name) +
				flowIDStyle.Render(f.ID) +
				sourceStyle.Render(f.Source) +
				colorRed.MaxWidth(30).Render(f.Desc)
		} else {
			s = iconOK.String() +
				flowNameStyle.Render(f.Name) +
				flowIDStyle.Render(f.ID) +
				sourceStyle.Render(f.Source) +
				lipgloss.NewStyle().MaxWidth(30).Render(f.Desc)
		}
		fmt.Fprintln(os.Stdout, s)
	}
	fmt.Fprintf(os.Stdout, "\n")
	return nil
}
