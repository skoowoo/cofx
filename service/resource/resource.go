package resource

import "io"

// Resources contains some services that can be used by the driver and function.
// .e.g. logset service, cron service, httpserver service etc.
type Resources struct {
	Logwriter io.Writer
}
