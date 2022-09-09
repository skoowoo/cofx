package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/cofunclabs/cofunc/service"
)

func listStd() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	all := svc.ListStdFunctions(ctx)

	// here is title
	fmt.Fprintln(os.Stdout, "\n"+
		colorGrey.Render(iconSpace.String()+
			flowNameStyle.Render("FUNCTION NAME")+
			"DESC"))

	for _, f := range all {
		s := iconCircleOk.String() +
			funcNameStyle.Render(f.Name) +
			lipgloss.NewStyle().MaxWidth(100).Render(f.Desc)
		fmt.Fprintln(os.Stdout, s)
	}
	fmt.Fprintf(os.Stdout, "\n")
	return nil
}

func inspectStd(fname string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	return svc.InspectStdFunction(ctx, fname).JsonWrite(os.Stdout)
}
