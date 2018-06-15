package runner

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/log"
	"github.com/coldog/bld/pkg/store"
	"github.com/stretchr/testify/require"
)

func test(t *testing.T, build builder.Build) {
	log.Level(4)

	tmp, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	r := &Runner{
		Build:    build,
		BuildDir: tmp,
		Store:    store.NewLocalStore(tmp),
		Perform: func(context.Context, builder.StepExec) error {
			return nil
		},
	}

	err = r.Run(context.Background())
	require.Nil(t, err)

	err = r.Run(context.Background())
	require.Nil(t, err)

	err = r.Run(context.Background())
	require.Nil(t, err)
}

func TestCases(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	dir2, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	err = ioutil.WriteFile(dir+"/test.txt", []byte("hello"), 0700)
	require.Nil(t, err)

	builds := map[string]builder.Build{
		"SingleDep": {
			Name: "test",
			ID:   "test-1",
			Sources: []builder.Source{
				{Name: "r1", Target: dir},
			},
			Steps: []builder.Step{{
				Name: "s1",
				Imports: []builder.Mount{
					{Source: "r1", Mount: "/usr/src/app"},
				},
				Exports: []builder.Mount{
					{Source: "r2", Mount: "/usr/src/app2"},
				},
			}},
		},
		"MultipleFromFirst": {
			Name: "test",
			ID:   "test-2",
			Sources: []builder.Source{
				{Name: "r1", Target: dir},
			},
			Steps: []builder.Step{
				{
					Name:    "s1-1",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
				{
					Name:    "s1-2",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
				{
					Name:    "s1-3",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
			},
		},
		"MultipleSources": {
			Name: "test",
			ID:   "test-3",
			Sources: []builder.Source{
				{Name: "r1", Target: dir},
				{Name: "r2", Target: dir2},
			},
			Steps: []builder.Step{
				{
					Name: "s1",
					Imports: []builder.Mount{
						{Source: "r1", Mount: "/usr/src/app"},
						{Source: "r1", Mount: "/usr/src/app2"},
					},
				},
			},
		},
	}

	for name, build := range builds {
		func(build builder.Build) {
			t.Run(name, func(t *testing.T) {
				// t.Parallel()
				test(t, build)
			})
		}(build)
	}
}
