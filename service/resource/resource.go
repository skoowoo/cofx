package resource

import (
	"context"
	"io"
	"net/http"
	"time"
)

type LogStdoutPrinter interface {
	PrintTitle()
	PrintSummary()
	Reset() error
}

// Resources contains some services that can be used by the driver and function.
// .e.g. logset service, cron service, httpserver service etc.
type Resources struct {
	Logwriter    io.Writer
	CronTrigger  CronTrigger
	HttpTrigger  HttpTrigger
	OutputParser TableOperation
	Labels       LabelManger
}

// LabelManager manage some labels for driver and function, the LabelManager is a resource.
type LabelManger interface {
	Get(key string) string
	Set(key, value string)
}

// CronTrigger add and remove the cron job by trigger function, the CronTrigger is a resource for trigger.
type CronTrigger interface {
	Add(format string, ch chan<- time.Time) (interface{}, error)
	Remove(interface{}) error
}

// HttpTrigger add and remove the http handler by trigger function, the HttpTrigger is a resource for trigger.
type HttpTrigger interface {
	AddRoute(path string, handler func(w http.ResponseWriter, r *http.Request)) error
	RemoveRoute(path string) error
}

// TableOperation is the interface for db table's insert, delete and query.
type TableOperation interface {
	Insert(ctx context.Context, columns []string, values ...any) error
	Delete(ctx context.Context, where string) error
	Query(ctx context.Context, columns []string, where string) ([][]string, error)
}
