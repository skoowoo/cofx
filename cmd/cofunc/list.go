package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
)

func listFlows(interactive bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables := svc.ListAvailables(ctx)

	// execute 'cofunc list' command
	if !interactive {
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, colorGrey.Render(iconSpace.String()+flowNameStyle.Render("FLOW NAME")+flowIDStyle.Render("FLOW ID")+"SOURCE"))
		for _, f := range availables {
			s := iconCircle.String() + flowNameStyle.Render(f.Name) + flowIDStyle.Render(f.ID) + f.Source
			fmt.Fprintln(os.Stdout, s)
		}
		fmt.Fprintln(os.Stdout, "")
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
			err := prunflowl(nameid.NameOrID(selected.Source), true)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}
