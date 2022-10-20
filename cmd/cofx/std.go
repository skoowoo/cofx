package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	pretty "github.com/cofxlabs/cofx/pkg/pretty"
	"github.com/cofxlabs/cofx/service"
)

func listStd() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	all := svc.ListStdFunctions(ctx)
	sort.Slice(all, func(i int, j int) bool {
		a1 := all[i].Category + "/" + all[i].Name
		a2 := all[j].Category + "/" + all[j].Name
		return a1 < a2
	})

	// here is title
	fmt.Fprintln(os.Stdout, "\n"+
		colorGrey.Render(iconSpace.String()+
			funcNameStyle.Render("FUNCTION NAME")+
			"DESC"))

	for _, f := range all {
		name := f.Name
		if f.Category != "" {
			name = f.Category + "/" + name
		}
		s := pretty.IconMinCircleOk.String() +
			funcNameStyle.Foreground(lipgloss.Color("222")).Render(name) +
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
