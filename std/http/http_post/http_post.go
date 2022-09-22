package httppost

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/tidwall/gjson"
)

var urlArg = manifest.UsageDesc{
	Name: "url",
	Desc: "Specify the url address to access",
}

var pathArg = manifest.UsageDesc{
	Name: "query_json_path",
	Desc: "Specify the path to get values from json document, the path is provided with GJSON Syntax:\nhttps://github.com/tidwall/gjson/blob/master/SYNTAX.md",
}

var headersArg = manifest.UsageDesc{
	Name: "set_headers",
	Desc: "Set some headers that you want to send to the server",
}

var cookiesArg = manifest.UsageDesc{
	Name: "set_cookies",
	Desc: "Set some cookies that you want to send to the server",
}

var jsonBodyArg = manifest.UsageDesc{
	Name: "json_file_path",
	Desc: "Specify the json file path to read as the http POST request body",
}

var _manifest = manifest.Manifest{
	Category:       "http",
	Name:           "http_post",
	Description:    "Send a http POST request to a service, and then handle the response",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{urlArg, pathArg, headersArg, cookiesArg, jsonBodyArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

type Custom struct {
	BodyReader io.ReadCloser
	// for testing
	client *http.Client
}

func (c *Custom) Close() error {
	if c.BodyReader != nil {
		return c.BodyReader.Close()
	}
	c.client = nil
	return nil
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, func() spec.Customer {
		return &Custom{}
	}
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	url, err := args.GetURL(urlArg.Name)
	if err != nil {
		return nil, err
	}
	headers := args.GetStringSlice(headersArg.Name)

	br, err := args.GetReader(jsonBodyArg.Name)
	if err != nil {
		return nil, err
	}
	if br == nil {
		br = bundle.Custom.(*Custom).BodyReader
	}
	defer func() {
		if br != nil {
			br.Close()
		}
	}()

	tr := &http.Transport{
		MaxIdleConns:       5,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	// Used to unit test
	if bundle.Custom.(*Custom).client != nil {
		client = bundle.Custom.(*Custom).client
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, br)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for _, s := range headers {
		fields := strings.Split(s, ":")
		if len(fields) == 2 {
			k := strings.TrimSpace(fields[0])
			v := strings.TrimSpace(fields[1])
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read body
	var buff bytes.Buffer
	io.Copy(&buff, resp.Body)

	// handle error
	if resp.StatusCode-200 >= 100 {
		return nil, fmt.Errorf("status code not 2xx: %d: %s", resp.StatusCode, buff.String())
	}

	// success
	returns := make(map[string]string)
	returns["status_code"] = strconv.Itoa(resp.StatusCode)

	paths := args.GetStringSlice(pathArg.Name)
	if len(paths) != 0 {
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "application/json") {
			results := gjson.GetManyBytes(buff.Bytes(), paths...)
			for i, res := range results {
				k := paths[i]
				v := res.String()
				returns[k] = v
			}
		}
	}
	return returns, nil
}
