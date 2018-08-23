package fileutils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	err := Copy("testdata", dir, []string{"test.txt"})
	require.Nil(t, err)
}
