package time

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

var _manifest = manifest.Manifest{
	Name:           "time",
	Description:    "Used to get the current time information",
	Driver:         "go",
	EntryPoint:     "",
	EntrypointFunc: Entrypoint,
	Args: map[string]string{
		"format":        "YYYY-MM-DD hh:mm:ss",
		"get_timestamp": "false",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{
			{
				Name: "format",
				OptionalValues: []string{
					"YYYY-MM-DD hh:mm:ss",
					"YYYY/MM/DD hh:mm:ss",
					"MM-DD-YYYY hh:mm:ss",
					"MM/DD/YYYY hh:mm:ss",
				},
				Desc: `Specifies the format for getting the current time`,
			},
			{
				Name: "get_timestamp",
				OptionalValues: []string{
					"true",
					"false",
				},
				Desc: `get timestamp in second`,
			},
		},
		ReturnValues: []manifest.UsageDesc{
			{
				Name: "now",
				Desc: "Current time",
			},
			{
				Name: "year",
				Desc: "The number of year",
			},
			{
				Name: "month",
				Desc: "Month",
			},
			{
				Name: "day",
				Desc: "The number of day",
			},
			{
				Name: "hour",
				Desc: "The number of hour",
			},
			{
				Name: "minute",
				Desc: "The number of minute",
			},
			{
				Name: "second",
				Desc: "The number of second",
			},
			{
				Name: "timestamp",
				Desc: "The number of timestamp",
			},
		},
	},
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
	format := args["format"]
	getts := args["get_timestamp"]

	var (
		now       string
		year      int
		month     string
		day       int
		hour      int
		minute    int
		second    int
		timestamp int64
	)

	current := time.Now()
	year = current.Year()
	month = current.Month().String()
	day = current.Day()
	hour = current.Hour()
	minute = current.Minute()
	second = current.Second()

	switch format {
	case "YYYY-MM-DD hh:mm:ss":
		now = current.Format("2006-01-02 15:04:05")
	case "YYYY/MM/DD hh:mm:ss":
		now = current.Format("2006/01/02 15:04:05")
	case "MM-DD-YYYY hh:mm:ss":
		now = current.Format("01-02-2006 15:04:05")
	case "MM/DD/YYYY hh:mm:ss":
		now = current.Format("01/02/2006 15:04:05")
	default:
		return nil, errors.New("invalid format argument: " + format)
	}

	ret := map[string]string{
		"now":    now,
		"year":   strconv.Itoa(year),
		"month":  month,
		"day":    strconv.Itoa(day),
		"hour":   strconv.Itoa(hour),
		"minute": strconv.Itoa(minute),
		"second": strconv.Itoa(second),
	}

	if getts == "true" {
		timestamp = current.Unix()
		ret["timestamp"] = fmt.Sprintf("%d", timestamp)
	}
	return ret, nil
}
