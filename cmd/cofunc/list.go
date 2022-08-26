package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cofunclabs/cofunc/service"
)

func listFlows(interactive bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables, err := svc.ListAvailableFlows(ctx)
	if err != nil {
		return err
	}

	// execute 'cofunc list' command
	if !interactive {
		fmt.Fprintln(os.Stdout, flowNameStyle.Render("FLOW NAME")+flowIDStyle.Render("FLOW ID")+"SOURCE")
		for _, f := range availables {
			s := flowNameStyle.Render(f.Name) + flowIDStyle.Render(f.ID) + f.Source
			fmt.Fprintln(os.Stdout, s)
		}
		return nil
	}

	//  execute 'cofunc' command without any args or sub-command
	for {
		selected, err := startListingView(availables)
		if err != nil {
			return err
		}

		// to run the selected flow
		if selected.Source != "" {
			err := runflowl(selected.Source, true)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}
