package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/executor"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/runner"
	"github.com/coldog/bld/pkg/store"
)

func exitErr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	var (
		target      string
		buildDir    string
		rootDir     string
		backend     string
		concurrency int
		level       uint
	)
	wd, _ := os.Getwd()

	flag.StringVar(&target, "target", wd, "target directory for the build")
	flag.StringVar(&buildDir, "build-dir", wd+"/.bld", "target directory for the build")
	flag.StringVar(&rootDir, "root-dir", wd, "root directory for the build")
	flag.StringVar(&backend, "backend", "local", "storage backend")
	flag.UintVar(&level, "v", 0, "log verbosity")
	flag.IntVar(&concurrency, "concurrency", 5, "maximum concurrency")
	flag.Parse()

	log.Level(uint32(level))

	var build builder.Build
	{
		file := target + "/.bld.json"
		f, err := os.Open(file)
		if err != nil {
			exitErr("Failed to open: %v", err)
		}
		err = json.NewDecoder(f).Decode(&build)
		if err != nil {
			exitErr("Failed to decode (%s): %v", file, err)
		}
	}

	build.ID = id()

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
