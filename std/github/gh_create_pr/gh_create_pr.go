package ghcreatepr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/manifest"
	"github.com/skoowoo/cofx/service/resource"
	"github.com/skoowoo/cofx/std/command"
	httppost "github.com/skoowoo/cofx/std/http/http_post"
)

var (
	urlArg = manifest.UsageDesc{
		Name: "create_pr_url",
		Desc: "Specify the url to create a pull request as a api",
	}
	fromBranchArg = manifest.UsageDesc{
		Name: "from_branch",
		Desc: "Specify the source branch name that will be merged",
	}
	toBranchArg = manifest.UsageDesc{
		Name: "to_branch",
		Desc: "Specify the target branch name that will merge into",
	}
	fromOrgArg = manifest.UsageDesc{
		Name: "from_org",
		Desc: "Specify the source org name, maybe it's a personal org",
	}
	tokenArg = manifest.UsageDesc{
		Name: "github_token",
		Desc: "Specify your github token",
	}
)

var (
	statusRet = manifest.UsageDesc{
		Name: "status_code",
		Desc: "Returns the status code of creating pull request",
	}
	prHtmlRet = manifest.UsageDesc{
		Name: "pr_html",
		Desc: "Returns the directly accessible pull request html",
	}
)

var _manifest = manifest.Manifest{
	Category:       "github",
	Name:           "gh_create_pr",
	Description:    "Create a pull request to upstream",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{urlArg, toBranchArg, fromBranchArg, fromOrgArg, tokenArg},
		ReturnValues: []manifest.UsageDesc{statusRet, prHtmlRet},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	toBranch, fromBranch := args.GetString(toBranchArg.Name), args.GetString(fromBranchArg.Name)
	if toBranch == "" || fromBranch == "" {
		return nil, fmt.Errorf("not specified %s or %s", toBranchArg.Name, fromBranchArg.Name)
	}
	// Figure out the commits information about the branch that will be merged
	_args := spec.EntrypointArgs{
		"cmd":            "git rev-list --left-right --pretty=oneline  " + fromBranch + "...upstream/" + toBranch,
		"split":          "--not-split-it--",
		"extract_fields": "0",
		"query_columns":  "c0",
		"query_where":    "c0 like '<%'",
	}
	_, ep, _ := command.New()
	rets, err := ep(ctx, bundle, _args)
	if err != nil {
		return nil, fmt.Errorf("%w: in %s function", err, _manifest.Name)
	}
	if len(rets) == 0 {
		return nil, fmt.Errorf("no commits for creating pull request")
	}
	// e.g. '<388f1f257be0e4e5b61db042fbdbf0423f271601 feature: added github auto-pr flow'
	var firstCommit string
	for k, v := range rets {
		if strings.HasSuffix(k, "_0") {
			firstCommit = v
			break
		}
	}
	if fields := strings.SplitN(firstCommit, " ", 2); len(fields) == 2 {
		firstCommit = fields[1]
	}

	// Create pull request by calling github api
	{
		fromOrg := args.GetString(fromOrgArg.Name)
		if fromOrg == "" {
			return nil, fmt.Errorf("not specified %s ", fromOrgArg.Name)
		}
		cpr := CreatePullRequest{
			Title: firstCommit,
			Body:  "",
			Head:  fromOrg + ":" + fromBranch,
			Base:  toBranch,
		}
		var buff bytes.Buffer
		json.NewEncoder(&buff).Encode(&cpr)
		fmt.Fprintf(bundle.Resources.Logwriter, "create pull request: %s\n", buff.String())

		url := args.GetString(urlArg.Name)
		if url == "" {
			return nil, fmt.Errorf("not specified " + urlArg.Name)
		}
		token := args.GetString(tokenArg.Name)
		if token == "" {
			return nil, fmt.Errorf("not specified " + tokenArg.Name)
		}

		_, post, custom := httppost.New()

		_bundle := spec.EntrypointBundle{
			Version: "latest",
			Custom:  custom(),
			Resources: resource.Resources{
				Logwriter: bundle.Resources.Logwriter,
				Labels:    bundle.Resources.Labels,
			},
		}
		_bundle.Custom.(*httppost.Custom).BodyReader = bytes.NewReader(buff.Bytes())
		args := spec.EntrypointArgs{
			"url":             url,
			"set_headers":     "Accept: application/vnd.github+json, Authorization: Bearer " + token,
			"query_json_path": "_links.html.href",
		}
		rets, err := post(ctx, _bundle, args)
		if err != nil {
			return nil, fmt.Errorf("%w: in %s function", err, _manifest.Name)
		}
		rets[prHtmlRet.Name] = rets["_links.html.href"]
		delete(rets, "_links.html.href")
		return rets, nil
	}
}

type CreatePullRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}
