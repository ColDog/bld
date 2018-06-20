package content

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDigestDir(t *testing.T) {
	digest, err := DigestDir("../../vendor")
	require.Nil(t, err)
	require.NotEqual(t, "", digest)
	println(digest)
}

func TestDigestFiles(t *testing.T) {
	digest, err := DigestFiles("../../cmd", []string{"bld/main.go"})
	require.Nil(t, err)
	require.NotEqual(t, "", digest)
	println(digest)
}

func TestDigestStrings(t *testing.T) {
	digest := DigestStrings("test1", "test2")
	require.NotEqual(t, "", digest)
	println(digest)
}
