package store

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreLocal_Save(t *testing.T) {
	sDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	s := &local{dir: sDir}

	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	err = ioutil.WriteFile(dir+"/test.txt", []byte("hello"), 0700)
	require.Nil(t, err)

	err = s.Save("test", dir)
	require.Nil(t, err)

	loadDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	err = s.Load("test", loadDir)
	require.Nil(t, err)

	data, err := ioutil.ReadFile(loadDir + "/test.txt")
	require.Nil(t, err)
	require.Equal(t, string(data), "hello")
}

func TestStoreLocal_Keys(t *testing.T) {
	sDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	s := &local{dir: sDir}

	err = s.PutKey("test", "hi")
	require.Nil(t, err)

	val, err := s.GetKey("test")
	require.Nil(t, err)
	require.Equal(t, val, "hi")
}
