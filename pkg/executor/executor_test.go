package executor

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/coldog/bld/pkg/fileutils"
	"github.com/coldog/bld/pkg/genid"

	"github.com/coldog/bld/pkg/builder"
	"github.com/coldog/bld/pkg/log"
	"github.com/stretchr/testify/require"
)

func test(t *testing.T, step builder.StepExec) {
	log.Level(4)

	e := &Executor{}

	err := e.Open()
	require.Nil(t, err)

	tmp, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	step.BuildDir = tmp

	err = e.Execute(context.Background(), step)
	require.Nil(t, err)
}

func TestEcho(t *testing.T) {
	test(t, builder.StepExec{
		Step: builder.Step{
			Name:     "test",
			Image:    "alpine",
			Commands: []string{"echo 'hello'"},
		},
		BuildID: genid.ID(),
	})
}

func TestMount(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	cacheTmp, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	err = ioutil.WriteFile(tmp+"/test.txt", []byte("I am in a file!"), fileutils.Regular)
	require.Nil(t, err)

	test(t, builder.StepExec{
		Step: builder.Step{
			Name:     "test",
			Image:    "alpine",
			Commands: []string{"echo 'hello'", "cat /mnt/test.txt"},
			Imports:  []builder.Mount{{Source: "test", Mount: "/mnt"}},
			Volumes:  []builder.Mount{{Source: "cache", Mount: "/cache"}},
		},
		BuildID: genid.ID(),
		SourceDirs: map[string]string{
			"test":  tmp,
			"cache": cacheTmp,
		},
	})
}

func TestMountExport(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	test(t, builder.StepExec{
		Step: builder.Step{
			Name:     "test",
			Image:    "alpine",
			Commands: []string{"echo 'hello' > /mnt/test.txt"},
			Exports:  []builder.Mount{{Source: "out", Mount: "/mnt"}},
		},
		BuildID: genid.ID(),
		SourceDirs: map[string]string{
			"out": tmp,
		},
	})

	data, err := ioutil.ReadFile(tmp + "/test.txt")
	require.Nil(t, err)
	require.Equal(t, "hello\n", string(data))
}

func TestCommit(t *testing.T) {
	test(t, builder.StepExec{
		Step: builder.Step{
			Name:     "test",
			Image:    "alpine",
			Commands: []string{"echo 'hello' > /test.txt"},
			Save: builder.Image{
				Tag:        "test",
				Entrypoint: []string{"/bin/sh"},
			},
		},
		BuildID: genid.ID(),
	})
}
