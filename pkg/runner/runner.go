package runner

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/coldog/bld/pkg/fileutils"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/content"
	"github.com/coldog/bld/pkg/graph"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/store"
)

// Runner will run a specific build.
type Runner struct {
	Store    store.Store
	BuildDir string
	RootDir  string
	Build    builder.Build
	Perform  func(ctx context.Context, step builder.StepExec) error
	Workers  int

	steps map[string]string

	logger        log.Logger
	lock          sync.RWMutex
	sourceDirs    map[string]string
	sourceDigests map[string]string
}

func (r *Runner) recordStep(name, digest string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.steps == nil {
		r.steps = map[string]string{}
	}
	r.steps[name] = digest
}

// AddSrc goes through the workflow of adding a source directory.
func (r *Runner) addSrc(name, target string, files []string) error {
	if err := os.MkdirAll(target, fileutils.Directory); err != nil {
		r.logger.V(4).Printf("failed to mkdirall target dir: %v", err)
	}

	var digest string
	{
		var err error
		if len(files) > 0 {
			digest, err = content.DigestFiles(target, files)
			if err != nil {
				return err
			}
		} else {
			digest, err = content.DigestDir(target)
			if err != nil {
				return err
			}
		}

	}

	r.logger.V(4).Printf("adding source src=%s target=%s digest=%s", name, target, digest)
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.sourceDigests == nil {
		r.sourceDigests = map[string]string{}
	}
	if r.sourceDirs == nil {
		r.sourceDirs = map[string]string{}
	}
	r.sourceDigests[name] = digest
	r.sourceDirs[name] = target
	return nil
}

func (r *Runner) getSrcDir(name string) string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.sourceDirs[name]
}

func (r *Runner) getSrcDigest(name string) string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.sourceDigests[name]
}

func (r *Runner) collectSources() map[string]string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	m := map[string]string{}
	r.logger.V(4).Printf("collected sources m=%+v", m)
	for name, dir := range r.sourceDirs {
		m[name] = dir
	}
	for _, vol := range r.Build.Volumes {
		m[vol.Name] = vol.Target
	}
	return m
}

func (r *Runner) dir(name string) string { return r.RootDir + "/" + name }

func (r *Runner) sourceDir(name string) string {
	return r.BuildDir + "/sources/" + r.Build.ID + "/" + name + "/"
}

// RunStep executes all instructions for a given step. The workflow is as
// follows:
// 1. Digest all imports.
// 2. Check if a digest exists from (1).
// If the digest exists:
// 	 3. Restore all exports from the found digest.
// If the digest does not exist:
//   3. Prepare exports for mounting.
//   4. Run the step and save all exports.
func (r *Runner) runStep(ctx context.Context, step builder.Step) error {
	start := time.Now()

	imports := []string{}
	for _, imp := range step.Imports {
		imports = append(imports, r.getSrcDigest(imp.Source))
	}
	digest := content.DigestStrings(imports...)
	r.recordStep(step.Name, digest)

	logger := r.logger.Prefix(r.Build.Name + "/" + step.Name)
	logger.Printf("STEP: %s (%s)", step.Name, digest)

	if _, err := r.Store.GetKey("step/" + r.Build.Name + "/" + step.Name + "/" + digest); err == nil {
		logger.V(4).Printf("restoring exports digest=%s step=%+v", digest, step)
		logger.Printf("> %s: step cached (%v)", step.Name, time.Since(start))
		return r.restoreExports(ctx, digest, step)
	}
	logger.V(4).Printf("running step digest=%s step=%+v", digest, step)

	if err := r.prepareExports(ctx, step); err != nil {
		return err
	}

	ctx = log.ContextWithLogger(ctx, logger)
	exec := builder.StepExec{
		Step:       step,
		SourceDirs: r.collectSources(),
		BuildDir:   r.BuildDir,
		BuildID:    r.Build.ID,
		RootDir:    r.RootDir,
	}
	logger.V(4).Printf("executing step: %+v", exec)
	if err := r.Perform(ctx, exec); err != nil {
		return err
	}

	logger.V(4).Printf("saving exports %+v", step.Exports)
	if err := r.saveExports(ctx, digest, step); err != nil {
		return err
	}

	logger.Printf("> %s: step finished (%v)", step.Name, time.Since(start))
	return r.Store.PutKey("step/"+r.Build.Name+"/"+step.Name+"/"+digest, "")
}

// RestoreExports will mount exports from the store.
func (r *Runner) restoreExports(
	ctx context.Context, digest string, step builder.Step) error {
	for _, exp := range step.Exports {
		var key string
		{
			var err error
			key, err = r.Store.GetKey("export/" + exp.Source + "/" + digest)
			if err != nil {
				return fmt.Errorf("failed to get export %s: %v", exp.Source, err)
			}
		}

		dir := r.sourceDir(exp.Source)
		if err := r.Store.Load(key, dir); err != nil {
			return fmt.Errorf("failed to load: %v", err)
		}
		if err := r.addSrc(exp.Source, dir, nil); err != nil {
			return fmt.Errorf("failed to restore %s: %v", exp.Source, err)
		}
	}
	return nil
}

// PrepareExports will setup directories for exports so that they can be
// mounted.
func (r *Runner) prepareExports(ctx context.Context, step builder.Step) error {
	for _, exp := range step.Exports {
		dir := r.sourceDir(exp.Source)
		if err := r.addSrc(exp.Source, dir, nil); err != nil {
			return fmt.Errorf("failed to prepare %s: %v", exp.Source, err)
		}
	}
	return nil
}

// SaveExports will select exports for a given step and save them using the
// store implementation.
func (r *Runner) saveExports(
	ctx context.Context, digest string, step builder.Step) error {
	for _, exp := range step.Exports {
		dir := r.sourceDir(exp.Source)
		if err := r.addSrc(exp.Source, dir, nil); err != nil {
			return err
		}
		sourceDigest := r.getSrcDigest(exp.Source)
		if err := r.Store.PutKey("export/"+exp.Source+"/"+digest, sourceDigest); err != nil {
			return fmt.Errorf("failed to get export %s: %v", exp.Source, err)
		}
		if err := r.Store.Save(sourceDigest, dir); err != nil {
			return err
		}
	}
	return nil
}

// RunSource will copy the source from the original target directory into a
// scratch source directory after it is checksummed.
// TODO: Performance improvement here is to only copy when changed.
func (r *Runner) runSource(ctx context.Context, src builder.Source) error {
	srcDir := r.dir(src.Target)
	destDir := r.sourceDir(src.Name)
	if err := fileutils.Copy(srcDir, destDir); err != nil {
		return err
	}
	return r.addSrc(src.Name, destDir, src.Files)
}

// Run will run a given target. It expects source targets to match:
// `source/name`.
func (r *Runner) run(ctx context.Context, name string) error {
	sourceName, ok := isSource(name)
	if ok {
		source, ok := r.Build.Source(sourceName)
		if !ok {
			return fmt.Errorf("source not found: %s", name)
		}
		return r.runSource(ctx, source)
	}
	step, ok := r.Build.Step(name)
	if !ok {
		return fmt.Errorf("step not found: %s", name)
	}
	return r.runStep(ctx, step)
}

// Checksum returns the checksum for the entire build, it depends on the step
// map populated by each step.
func (r *Runner) checksum() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	keys := []string{}
	for k := range r.steps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	digests := []string{}
	for _, k := range keys {
		digests = append(digests, r.steps[k])
	}
	return content.DigestStrings(digests...)
}

// Run executes the build, it will exit if the context is closed.
func (r *Runner) Run(ctx context.Context) error {
	s := &graph.Solver{
		Build: r.Build,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := r.logger.Prefix(r.Build.Name)
	r.logger = log

	errs := make(chan error, r.Workers)
	wg := &sync.WaitGroup{}
	wg.Add(r.Workers)

	log.Printf("starting build %s", r.Build.ID)
	log.V(4).Printf("%+v", r.Build)

	s.Solve()

	for i := 0; i < r.Workers; i++ {
		go func(i int) {
			log := log.Prefix(fmt.Sprintf("%s/%d", r.Build.ID, i))
			defer wg.Done()
			defer log.V(4).Printf("worker exited %d", i)

			for {
				id, err := s.Select(ctx)
				if err == graph.ErrFinished {
					return
				}
				if err != nil {
					errs <- err
					return
				}
				log.V(4).Printf("starting step %s", id)
				err = r.run(ctx, id)
				if err != nil {
					log.V(4).Printf("step failed %s: %v", id, err)
					errs <- err
					return
				}
				log.V(4).Printf("finished step %s", id)
				s.Done(id)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(errs)
		s.Close()
	}()

	for e := range errs {
		return e
	}

	log.Printf("finished (%s)", r.checksum())
	return nil
}

func isSource(name string) (string, bool) {
	spl := strings.Split(name, "/")
	if spl[0] == "source" {
		return spl[1], true
	}
	return "", false
}
