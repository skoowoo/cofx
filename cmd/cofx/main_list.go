package main

import (
	"context"

	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/cofxlabs/cofx/service"
)

func mainList() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := service.New()
	availables := svc.ListAvailables(ctx)

	//  execute 'cofx' command without any args or sub-command
	for {
		selected, err := startListingView(availables)
		if err != nil {
			return err
		}

		// to run the selected flow
		if selected.Source != "" {
			err := prunEntry(nameid.NameOrID(selected.Source), true)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}
