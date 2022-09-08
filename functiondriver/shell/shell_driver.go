package shelldriver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cofunclabs/cofunc/config"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/pkg/output"
	"github.com/cofunclabs/cofunc/service/resource"
)

const Name = "shell"

// ShellDriver is used to execute shell script functions. All shell script functions must be stored in
// $COFUNC_HOME/shell directory, the ShellDriver is able to find and load them.
type ShellDriver struct {
	fpath   string
	fname   string
	version string
	// manifest be defined by function
	manifest *manifest.Manifest
	// resources contains some services that can be used by driver self. the shell function
	// inability to use any service.
	resources resource.Resources
}

// New creates a new ShellDriver instance to execute shell script functions.
func New(fname, fpath, version string) *ShellDriver {
	return &ShellDriver{
		fname:   fname,
		fpath:   fpath,
		version: version,
	}
}

// Load loads the shell script function from $COFUNC_HOME/shell directory.
func (d *ShellDriver) Load(ctx context.Context, resources resource.Resources) error {
	functionDir := filepath.Join(config.ShellDir(), d.fpath)
	mfPath := filepath.Join(functionDir, "manifest.json")
	file, err := os.Open(mfPath)
	if err != nil {
		return fmt.Errorf("%w: shell driver load", err)
	}
	var _manifest manifest.Manifest
	if err := json.NewDecoder(file).Decode(&_manifest); err != nil {
		return fmt.Errorf("%w: shell driver decode manifest", err)
	}

	if _manifest.Entrypoint == "" {
		return fmt.Errorf("not found entrypoint in shell function: %s", d.fname)
	}
	program := filepath.Join(functionDir, _manifest.Entrypoint)
	if _, err := os.Stat(program); os.IsNotExist(err) {
		return fmt.Errorf("%w: not found entrypoint program", err)
	}

	d.manifest = &_manifest
	d.resources = resources

	return nil
}

// Run executes the shell script function, Please note that 'args' will be converted to environment
func (d *ShellDriver) Run(ctx context.Context, args map[string]string) (map[string]string, error) {
	defer func() {
		reseter, ok := d.resources.Logwriter.(resource.LogWriteReseter)
		if ok {
			reseter.Reset()
		}
	}()
	merged := d.mergeArgs(args)
	functionDir := filepath.Join(config.ShellDir(), d.fpath)
	program := filepath.Join(functionDir, d.manifest.Entrypoint)

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", program)
	cmd.Dir = functionDir
	cmd.Env = append(cmd.Env, d.toEnv(merged)...)

	retValues := make(map[string]string)
	out := &output.Output{
		W: d.resources.Logwriter,
		HandleFunc: func(line []byte) {
			// TODO:
		},
	}
	cmd.Stderr = out
	cmd.Stdout = out
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return retValues, nil
}

// StopAndReslease is used to stop and release the all resources.
func (d *ShellDriver) StopAndRelease(ctx context.Context) error {
	return nil
}

// FunctionName returns the name of the shell script function.
func (d *ShellDriver) FunctionName() string {
	return d.fname
}

// Name returns the name of the shell driver.
func (d *ShellDriver) Name() string {
	return Name
}

// Manifest returns the manifest of the shell script function.
func (d *ShellDriver) Manifest() manifest.Manifest {
	return *d.manifest
}

func (d *ShellDriver) mergeArgs(args map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range d.manifest.Args {
		merged[k] = v
	}
	for k, v := range args {
		merged[k] = v
	}
	return merged
}

func (d *ShellDriver) toEnv(args map[string]string) []string {
	var envs []string
	for k, v := range args {
		k = "COFUNC_" + strings.ToUpper(k)
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}
