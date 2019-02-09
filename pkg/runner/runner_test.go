package runner

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/store"
	"github.com/stretchr/testify/require"
)

type mockImageStore struct{}

func (mockImageStore) Save(ctx context.Context, name, id string) error    { return nil }
func (mockImageStore) Restore(ctx context.Context, name, id string) error { return nil }

var (
	wd   string
	tmp  string
	noop = func(ctx context.Context, exec builder.StepExec) error {
		return nil
	}
	fail = func(ctx context.Context, exec builder.StepExec) error {
		return errors.New("some err")
	}
)

func init() {
	wd, _ = os.Getwd()
	tmp, _ = ioutil.TempDir("", "")
}

func test(
	t *testing.T, build builder.Build,
	fn func(ctx context.Context, exec builder.StepExec) error) error {

	log.Level(4)

	r := &Runner{
		ImageStore: mockImageStore{},
		Store:      store.NewLocalStore(tmp),
		BuildDir:   tmp,
		RootDir:    wd,
		Build:      build,
		Workers:    2,
		Perform:    fn,
	}
	return r.Run(context.Background())
}

func TestRunner(t *testing.T) {
	err := test(t, builder.Build{
		ID:   "10",
		Name: "test",
		Sources: []builder.Source{
			{Name: "test+r1", Target: "testdata"},
		},
		Steps: []builder.Step{
			{
				Name:    "s1",
				Imports: []builder.Mount{{Source: "test+r1", Mount: "/usr/src/app"}},
				Exports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app2"}},
			},
			{
				Name:    "s2",
				Imports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app"}},
			},
		},
	}, noop)
	require.Nil(t, err)
}

func TestRunnerCached(t *testing.T) {
	err := test(t, builder.Build{
		ID:   "10",
		Name: "test",
		Sources: []builder.Source{
			{Name: "r1", Target: "testdata"},
		},
		Steps: []builder.Step{
			{
				Name:    "s1",
				Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				Exports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app2"}},
			},
			{
				Name:    "s2",
				Imports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app"}},
			},
		},
	}, noop)
	require.Nil(t, err)
}

func TestRunnerFailing(t *testing.T) {
	err := test(t, builder.Build{
		ID:   "10",
		Name: "test-123",
		Sources: []builder.Source{
			{Name: "r1", Target: "testdata"},
		},
		Steps: []builder.Step{
			{
				Name:    "s11",
				Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				Exports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app2"}},
			},
			{
				Name:    "s22",
				Imports: []builder.Mount{{Source: "r2", Mount: "/usr/src/app"}},
			},
		},
	}, fail)
	require.Error(t, err)
}
