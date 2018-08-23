package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/executor"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/runner"
	"github.com/coldog/bld/pkg/store"
	uuid "github.com/satori/go.uuid"
)

func exitErr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	var (
		repoURL     string
		repoAuth    string
		buildDir    string
		buildSpec   string
		rootDir     string
		backend     string
		concurrency int
		level       uint
	)
	wd, _ := os.Getwd()

	flag.StringVar(&repoURL, "repo", "localhost:5000", "docker repo url")
	flag.StringVar(&repoAuth, "repo-auth", "eyJ1c2VybmFtZSI6ImRvY2tlciJ9Cg==", "docker repo auth")
	flag.StringVar(&buildSpec, "spec", wd+"/.bld.yaml", "build specification")
	flag.StringVar(&buildDir, "build-dir", "/tmp/bld", "target directory for the build")
	flag.StringVar(&rootDir, "root-dir", wd, "root directory for the build")
	flag.StringVar(&backend, "backend", "local", "storage backend")
	flag.UintVar(&level, "v", 0, "log verbosity")
	flag.IntVar(&concurrency, "concurrency", 5, "maximum concurrency")
	flag.Parse()

	log.Level(uint32(level))

	var build builder.Build
	{
		b, err := builder.Read(buildSpec)
		if err != nil {
			exitErr("Failed to read (%s): %v", buildSpec, err)
		}
		build = b
	}

	build.ID = uuid.NewV4().String()

	var s store.Store
	switch backend {
	case "local":
		s = store.NewLocalStore(buildDir)
	default:
		exitErr("Invalid store %s", backend)
	}

	var imageStore store.ImageStore
	{
		is, err := store.NewImageStore(repoURL, repoAuth)
		if err != nil {
			exitErr("Invalid repo store %s", err)
		}
		imageStore = is
	}

	e := &executor.Executor{}
	if err := e.Open(); err != nil {
		exitErr("Failed to initialize executor: %v", err)
	}

	r := &runner.Runner{
		Store:      s,
		ImageStore: imageStore,
		BuildDir:   buildDir,
		RootDir:    rootDir,
		Build:      build,
		Perform:    e.Execute,
		Workers:    concurrency,
	}

	if err := r.Run(context.Background()); err != nil {
		exitErr("Run failed: %v", err)
	}
}
