package runner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/content"
	"github.com/coldog/bld/pkg/graph"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/store"
)

// Runner is a runner.
type Runner struct {
	Build       builder.Build
	Store       store.Store
	BuildDir    string
	RootDir     string
	Concurrency int
	Perform     func(ctx context.Context, s builder.StepExec) error

	log     log.Logger
	sources *sources
}

// Run the build.
func (r *Runner) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := r.log.Prefix(r.Build.ID)
	start := time.Now()

	if err := r.init(); err != nil {
		return err
	}

	log.Printf("Starting build c=%v id=%s", r.Concurrency, r.Build.ID)

	g := graph.Solver{
		Build: r.Build,
	}
	g.Solve()
	defer g.Close()

	errC := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(r.Concurrency)

	for i := 0; i < r.Concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				key, err := g.Select(ctx)
				if err != nil {
					return
				}
				err = r.execute(ctx, key)
				if err != nil {
					errC <- err
					return
				}
				g.Done(key)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errC)
	}()

	var err error
	for e := range errC {
		err = e
		return e
	}
	log.Printf("Finished build in %v", time.Since(start))
	return err
}

func (r *Runner) init() error {
	r.sources = newSources()
	if r.Concurrency == 0 {
		r.Concurrency = 5
	}
	for _, vol := range r.Build.Volumes {
		r.sources.setDir(vol.Name, vol.Target)
	}
	return nil
}

func (r *Runner) getSourceDir(digest string) string {
	return r.BuildDir + "/content/" + digest
}

func (r *Runner) getWorkDir(name string) string {
	return r.BuildDir + "/build/" + r.Build.ID + "/exports/" + name
}

func (r *Runner) execute(ctx context.Context, key string) error {
	log := r.log.Prefix(r.Build.ID + "/" + key)
	log.V(4).Printf("Executing key=%s", key)

	start := time.Now()
	isSource := strings.HasPrefix(key, "source/")
	name := strings.Replace(key, "source/", "", 1)

	log.Printf("STEP: %s", key)

	if isSource {
		sourceDigest, err := r.buildSource(key, name)
		log.V(4).Printf("Built source name=%s digest=%s", name, sourceDigest)
		if err != nil {
			return err
		}
		log.Printf("--> %s in %v", sourceDigest, time.Since(start))
		return nil
	}

	step, ok := r.Build.Step(key)
	if !ok {
		return fmt.Errorf("invalid step: %s", key)
	}

	digest := r.stepDigest(ctx, log, step)
	log.V(4).Printf("step digest calculated digest=%s", digest)
	defer func() { log.Printf("--> %s in %v", digest, time.Since(start)) }()

	hasRun := r.stepHasRun(step.Name, digest)

	if hasRun {
		return r.restoreStep(ctx, log, digest, step)
	}
	return r.runStep(ctx, log, digest, step)
}

func (r *Runner) buildSource(key, name string) (string, error) {
	src, ok := r.Build.Source(name)
	if !ok {
		return "", fmt.Errorf("invalid source: %s", key)
	}
	sourceDigest, err := content.DigestDir(r.RootDir + "/" + src.Target)
	r.sources.setDigest(src.Name, sourceDigest)
	r.sources.setDir(src.Name, r.RootDir+"/"+src.Target)
	return sourceDigest, err
}

func (r *Runner) stepDigest(ctx context.Context, log log.Logger, step builder.Step) string {
	var strings []string
	for _, imp := range step.Imports {
		sourceDigest := r.sources.getDigest(imp.Source)
		log.V(5).Printf("Checking step source digest=%s dir=%s", sourceDigest, r.sources.getDir(imp.Source))
		if sourceDigest == "" {
			// Should never hit this.
			panic("digest not found " + imp.Source)
		}
		strings = append(strings, sourceDigest)
	}
	return content.DigestStrings(strings...)
}

func (r *Runner) restoreStep(ctx context.Context, log log.Logger, digest string, step builder.Step) error {
	for _, exp := range step.Exports {
		log.V(4).Printf("restore export digest=%s name=%s", digest, exp.Source)
		if err := r.restoreExport(log, digest, exp.Source); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runStep(ctx context.Context, log log.Logger, digest string, step builder.Step) error {
	// Setup working directories for exports.
	for _, exp := range step.Exports {
		sourceDir := r.getWorkDir(exp.Source)
		if err := os.MkdirAll(sourceDir, 0700); err != nil {
			return err
		}
		r.sources.setDir(exp.Source, sourceDir)
	}

	stepExec := builder.StepExec{
		Step:       step,
		BuildDir:   r.BuildDir,
		BuildID:    r.Build.ID,
		SourceDirs: r.sources.dirMap(),
	}

	// DO STEP!
	log.V(4).Printf("performing step name=%s", step.Name)
	if err := r.Perform(ctx, stepExec); err != nil {
		return err
	}

	// Setup working directories for exports.
	for _, exp := range step.Exports {
		log.V(4).Printf("saving export digest=%s name=%s", digest, exp.Source)

		if err := r.saveExport(digest, exp.Source); err != nil {
			return err
		}
	}
	return r.stepSave(step.Name, digest)
}

func (r *Runner) stepSave(name, buildDigest string) error {
	err := r.Store.PutKey("build/"+r.Build.Name+"/"+name+"/"+buildDigest, r.Build.ID)
	return err
}

func (r *Runner) stepHasRun(name, buildDigest string) bool {
	_, err := r.Store.GetKey("build/" + r.Build.Name + "/" + name + "/" + buildDigest)
	return err == nil
}

// Load source, given a previous build digest.
func (r *Runner) restoreExport(log log.Logger, buildDigest, name string) error {
	// Fetch the layer for the given build digest.
	var sourceDigest string
	{
		var err error
		sourceDigest, err = r.Store.GetKey("source/" + buildDigest + "/" + name)
		if err != nil {
			return err
		}
	}

	dir := r.getSourceDir(sourceDigest)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Load the layer into the directory.
	if err := r.Store.Load(sourceDigest, dir); err != nil {
		return err
	}

	// Set variables.
	log.V(5).Printf("Set fields source=%s digest=%s dir=%s", name, sourceDigest, dir)
	r.sources.setDigest(name, sourceDigest)
	r.sources.setDir(name, dir)
	return nil
}

func (r *Runner) saveExport(buildDigest, name string) error {
	dir := r.sources.getDir(name)
	if dir == "" {
		return fmt.Errorf("export not found: %s", name)
	}

	sourceDigest, err := content.DigestDir(dir)
	if err != nil {
		return err
	}

	// Save the layer.
	if err := r.Store.Save(sourceDigest, dir); err != nil {
		return err
	}

	if err := r.Store.PutKey("source/"+buildDigest+"/"+name, sourceDigest); err != nil {
		return err
	}
	return nil
}
