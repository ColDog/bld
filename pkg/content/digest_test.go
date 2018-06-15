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
