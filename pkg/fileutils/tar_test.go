package fileutils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTar(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.Close()
	file := f.Name()
	dir, _ := ioutil.TempDir("", "")

	t.Run("Tar", func(t *testing.T) {
		err := Tar("testdata", file)
		require.NoError(t, err)
	})

	t.Run("Untar", func(t *testing.T) {
		err := Untar(file, dir)
		require.NoError(t, err)
	})
}
