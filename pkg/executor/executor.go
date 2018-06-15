package executor

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/fileutils"
	"github.com/coldog/bld/pkg/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/moby/client"
)

const workspaceDir = "/.bld/workspace"

// Executor executes the build steps.
type Executor struct {
	logger log.Logger
	client *client.Client
}

// Open will initialize the executor and open a docker client.
func (e *Executor) Open() error {
	if e.client != nil {
		return nil
	}

	client, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// Sanity check.
	_, err = client.ContainerList(
		context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	e.client = client
	return nil
}

func (e *Executor) pullImage(ctx context.Context, image string) error {
	if _, _, err := e.client.ImageInspectWithRaw(ctx, image); err == nil {
		return nil
	}
	r, err := e.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer r.Close()
	io.Copy(ioutil.Discard, r)
	return nil
}

func (e *Executor) execDir(step builder.StepExec) string {
	return step.BuildDir + "/build/" + step.BuildID + "/workspace"
}

func (e *Executor) getBinds(step builder.StepExec) []string {
	execDir := e.execDir(step)
	binds := []string{
		execDir + ":" + workspaceDir,
	}
	for _, imp := range step.Imports {
		binds = append(binds, step.SourceDirs[imp.Source]+":"+imp.Mount)
	}
	for _, exp := range step.Exports {
		binds = append(binds, step.SourceDirs[exp.Source]+":"+exp.Mount)
	}
	for _, v := range step.Volumes {
		binds = append(binds, step.SourceDirs[v.Source]+":"+v.Mount)
	}
	return binds
}

func (e *Executor) getConfig(step builder.StepExec) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
	binds := e.getBinds(step)
	entrypoint := e.entrypointFile(step)

	config := &container.Config{
		Image: step.Image,
		Entrypoint: strslice.StrSlice{
			workspaceDir + "/" + entrypoint,
		},
		WorkingDir: step.Workdir,
		Env:        step.Env,
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}
	netConfig := &network.NetworkingConfig{}
	return config, hostConfig, netConfig
}

func (e *Executor) startContainer(ctx context.Context, step builder.StepExec, config *container.Config, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig) (string, error) {
	ct, err := e.client.ContainerCreate(
		ctx, config, hostConfig, netConfig, step.BuildID+"_"+step.Name)
	if err != nil {
		return "", err
	}
	if err := e.client.ContainerStart(
		ctx, ct.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return ct.ID, nil
}

func (e *Executor) waitForExit(ctx context.Context, id string) (int, error) {
	if _, err := e.client.ContainerWait(ctx, id); err != nil {
		return 0, err
	}
	inspect, err := e.client.ContainerInspect(ctx, id)
	if err != nil {
		return 0, err
	}
	if err := e.client.ContainerRemove(
		ctx, id, types.ContainerRemoveOptions{}); err != nil {
		return 0, err
	}
	return inspect.State.ExitCode, nil
}

func (e *Executor) entrypointFile(step builder.StepExec) string {
	return step.Name + "_step.sh"
}

// Execute will execute the provided step.
func (e *Executor) Execute(ctx context.Context, step builder.StepExec) error {
	execDir := e.execDir(step)
	entrypoint := e.entrypointFile(step)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e.logger.Printf("pulling %s", step.Image)
	if err := e.pullImage(ctx, step.Image); err != nil {
		return err
	}

	e.logger.V(4).Printf("building entrypoint entrypoint=%s", entrypoint)
	if err := buildEntrypoint(execDir+"/"+entrypoint, step.Commands); err != nil {
		return err
	}

	config, hostConfig, netConfig := e.getConfig(step)

	var id string
	e.logger.V(4).Printf("creating container name=%v container=%+v host=%+v",
		step.BuildID+"_"+step.Name, config, hostConfig)
	{
		var err error
		id, err = e.startContainer(ctx, step, config, hostConfig, netConfig)
		if err != nil {
			return err
		}
	}

	e.logger.V(4).Printf("container started id=%s", id)
	go e.logs(ctx, e.logger, id)

	var exitCode int
	{
		var err error
		exitCode, err = e.waitForExit(ctx, id)
		if err != nil {
			return err
		}
	}

	e.logger.V(4).Printf("container finished code=%v", exitCode)
	if exitCode != 0 {
		return fmt.Errorf("container: exit code %d", exitCode)
	}
	return nil
}

func (e *Executor) logs(
	ctx context.Context, l log.Logger, id string) error {
	reader, err := e.client.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		Follow:     true,
		ShowStderr: true,
		ShowStdout: true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()
	w := &dockerWriter{id: id, l: l}
	_, err = stdcopy.StdCopy(w, w, reader)
	return err
}

type dockerWriter struct {
	id string
	l  log.Logger
}

func (d *dockerWriter) Write(b []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		d.l.Printf("[%s] %s", d.id[:7], line)
	}
	return len(b), nil
}

func buildEntrypoint(file string, commands []string) error {
	if err := os.MkdirAll(filepath.Dir(file), fileutils.Directory); err != nil {
		return err
	}
	entrypoint, err := os.OpenFile(
		file, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fileutils.Executable)
	if err != nil {
		return fmt.Errorf("executor: could not build entrypoint: %v", err)
	}
	defer entrypoint.Close()
	entrypoint.WriteString("#!/bin/sh\n")
	for _, cmd := range commands {
		entrypoint.WriteString(cmd + "\n")
	}
	return nil
}
