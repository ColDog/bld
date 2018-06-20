package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/executor"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/runner"
	"github.com/coldog/bld/pkg/store"
	"github.com/ghodss/yaml"
	uuid "github.com/satori/go.uuid"
)

func exitErr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	var (
		target      string
		buildDir    string
		buildSpec   string
		rootDir     string
		backend     string
		concurrency int
		level       uint
	)
	wd, _ := os.Getwd()

	flag.StringVar(&target, "target", wd, "target directory for the build")
	flag.StringVar(&buildSpec, "spec", wd+"/.bld.yaml", "build specification")
	flag.StringVar(&buildDir, "build-dir", wd+"/.bld", "target directory for the build")
	flag.StringVar(&rootDir, "root-dir", wd, "root directory for the build")
	flag.StringVar(&backend, "backend", "local", "storage backend")
	flag.UintVar(&level, "v", 0, "log verbosity")
	flag.IntVar(&concurrency, "concurrency", 5, "maximum concurrency")
	flag.Parse()

	log.Level(uint32(level))

	var build builder.Build
	{
		f, err := os.Open(buildSpec)
		if err != nil {
			exitErr("Failed to open: %v", err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			exitErr("Failed to read: %v", err)
		}
		err = yaml.Unmarshal(data, &build)
		if err != nil {
			exitErr("Failed to decode (%s): %v", buildSpec, err)
		}
	}

	build.ID = uuid.NewV4().String()

	var s store.Store
	switch backend {
	case "local":
		s = store.NewLocalStore(buildDir)
	default:
		exitErr("Invalid store %s", backend)
	}

	e := &executor.Executor{}
	if err := e.Open(); err != nil {
		exitErr("Failed to initialize executor: %v", err)
	}

	r := &runner.Runner{
		Store:    s,
		BuildDir: buildDir,
		RootDir:  rootDir,
		Build:    build,
		Perform:  e.Execute,
		Workers:  concurrency,
	}

	if err := r.Run(context.Background()); err != nil {
		exitErr("Run failed: %v", err)
	}
}
