package time

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
)

var (
	formatArg = manifest.UsageDesc{
		Name: "format",
		OptionalValues: []string{
			"YYYY-MM-DD hh:mm:ss",
			"YYYY/MM/DD hh:mm:ss",
			"MM-DD-YYYY hh:mm:ss",
			"MM/DD/YYYY hh:mm:ss",
		},
		Desc: `Specifies the format for getting the current time, default YYYY-MM-DD hh:mm:ss`,
	}
	timestampArg = manifest.UsageDesc{
		Name: "get_timestamp",
		OptionalValues: []string{
			"true",
			"false",
		},
		Desc: `Get timestamp in second, default false`,
	}
)

var (
	nowRet = manifest.UsageDesc{
		Name: "now",
		Desc: "Current time",
	}
	yearRet = manifest.UsageDesc{
		Name: "year",
		Desc: "The number of year",
	}
	monthRet = manifest.UsageDesc{
		Name: "month",
		Desc: "Month",
	}
	dayRet = manifest.UsageDesc{
		Name: "day",
		Desc: "The number of day",
	}
	hourRet = manifest.UsageDesc{
		Name: "hour",
		Desc: "The number of hour",
	}
	minuteRet = manifest.UsageDesc{
		Name: "minute",
		Desc: "The number of minute",
	}
	secondRet = manifest.UsageDesc{
		Name: "second",
		Desc: "The number of second",
	}
	timestampRet = manifest.UsageDesc{
		Name: "timestamp",
		Desc: "The number of timestamp",
	}
)

var _manifest = manifest.Manifest{
	Name:        "time",
	Description: "Read the current time and return multiple time value related variables",
	Driver:      "go",
	Args: map[string]string{
		"format":        "YYYY-MM-DD hh:mm:ss",
		"get_timestamp": "false",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{formatArg, timestampArg},
		ReturnValues: []manifest.UsageDesc{nowRet, timestampRet, yearRet, monthRet, dayRet, hourRet, minuteRet, secondRet},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	format := args.GetString(formatArg.Name)
	getts, err := args.GetBool(timestampArg.Name)
	if err != nil {
		return nil, err
	}

	var (
		now       string
		timestamp int64
	)

	current := time.Now()
	year := current.Year()
	month := current.Month().String()
	day := current.Day()
	hour := current.Hour()
	minute := current.Minute()
	second := current.Second()

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
		nowRet.Name:    now,
		yearRet.Name:   strconv.Itoa(year),
		monthRet.Name:  month,
		dayRet.Name:    strconv.Itoa(day),
		hourRet.Name:   strconv.Itoa(hour),
		minuteRet.Name: strconv.Itoa(minute),
		secondRet.Name: strconv.Itoa(second),
	}

	if getts {
		timestamp = current.Unix()
		ret[timestampRet.Name] = fmt.Sprintf("%d", timestamp)
	}
	return ret, nil
}
