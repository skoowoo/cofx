package httpget

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
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

var _manifest = manifest.Manifest{
	Category:       "http",
	Name:           "http_get",
	Description:    "Send a HTTP GET request",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{urlArg, pathArg, headersArg, cookiesArg},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	url, err := args.GetURL(urlArg.Name)
	if err != nil {
		return nil, err
	}
	paths := args.GetStringSlice(pathArg.Name)
	if url == "" || len(paths) == 0 {
		return nil, fmt.Errorf("not specify the url or path argument")
	}

	tr := &http.Transport{
		MaxIdleConns:       5,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
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

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		results := gjson.GetManyBytes(buff.Bytes(), paths...)
		for i, res := range results {
			k := paths[i]
			v := res.String()
			returns[k] = v
		}
	}
	return returns, nil
}
