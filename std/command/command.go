package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/pkg/output"
)

var cmdArg = manifest.UsageDesc{
	Name: "cmd",
	Desc: "Specify a command to run",
}

var envArg = manifest.UsageDesc{
	Name: "env",
	Desc: "Specify environment variables for the command",
}

var workingDirArg = manifest.UsageDesc{
	Name: "working_dir",
	Desc: "Specify working directory for the command",
}

var splitArg = manifest.UsageDesc{
	Name: "split",
	Desc: "Specify a separator to split",
}

var extractArg = manifest.UsageDesc{
	Name: "extract_fields",
	Desc: "Specify one column or more to extract, e.g. 0,1,2",
}

var queryColumnArg = manifest.UsageDesc{
	Name: "query_columns",
	Desc: "Specify column names that you wanted to query",
}

var queryWhereArg = manifest.UsageDesc{
	Name: "query_where",
	Desc: "Specify where clause for query",
}

var _manifest = manifest.Manifest{
	Name:           "command",
	Description:    "Used to run a command",
	Driver:         "go",
	Entrypoint:     "",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	IgnoreFailure:  false,
	Usage: manifest.Usage{
		Args: []manifest.UsageDesc{cmdArg, envArg, workingDirArg, splitArg, queryColumnArg, queryWhereArg},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	// cmd
	cmdstr := args.GetString(cmdArg.Name)
	if cmdstr == "" {
		return nil, errors.New("command function miss argument: " + cmdArg.Name)
	}
	// env
	env := args.GetStringSlice(envArg.Name)
	// working dir
	workingDir := args.GetString(workingDirArg.Name)
	if workingDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workingDir = dir
	}

	// split -> extract -> query
	splitSep := args.GetString(splitArg.Name)
	extractFields, err := args.GetIntSlice(extractArg.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: extract %s", err, args.GetString(extractArg.Name))
	}
	queryColumns := args.GetStringSlice(queryColumnArg.Name)
	queryWhere := args.GetString(queryWhereArg.Name)

	flowId := bundle.Resources.Labels.Get("flow_id")
	nodeSeq := bundle.Resources.Labels.Get("node_seq")
	nodeName := bundle.Resources.Labels.Get("node_name")

	// user defer to delete db data
	defer func() {
		where := fmt.Sprintf("flow_id = '%s' AND node_seq = '%s'", flowId, nodeSeq)
		if err := bundle.Resources.OutputParser.Delete(ctx, where); err != nil {
			log.Println(fmt.Errorf("%w: delete command output", err))
		}
	}()

	var buff bytes.Buffer
	out := &output.Output{
		W: &buff,
		HandleFunc: output.ColumnFunc(splitSep, func(columns []string) {
			// insert db
			names := []string{"flow_id", "node_seq", "node_name"}
			values := []interface{}{flowId, nodeSeq, nodeName}
			for i, v := range columns {
				names = append(names, fmt.Sprintf("c%d", i))
				values = append(values, v)
			}
			if err := bundle.Resources.OutputParser.Insert(ctx, names, values...); err != nil {
				log.Println(fmt.Errorf("%w: insert command output", err))
			}
		}, extractFields...),
	}

	// start the command
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdstr)
	cmd.Env = append(cmd.Env, env...)
	cmd.Dir = workingDir

	fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())

	opipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	epipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// copy stdout and stderr to out object
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer out.Close()
		io.Copy(out, opipe)
		io.Copy(out, epipe)
	}()
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(bundle.Resources.Logwriter, "%s", buff.String())
		return nil, err
	}

	// query outcome
	rows, err := bundle.Resources.OutputParser.Query(ctx, queryColumns, queryWhere)
	if err != nil {
		return nil, err
	}
	returns := make(map[string]string)
	for i, r := range rows {
		if splitSep == "" {
			splitSep = " "
		}
		k := fmt.Sprintf("outcome_%d", i)
		v := strings.Join(r, splitSep)
		returns[k] = v
	}
	return returns, nil
}
