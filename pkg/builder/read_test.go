package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	b, err := Read("testdata/.bld.yaml")
	require.NoError(t, err)

	t.Run("StepExists", func(t *testing.T) {
		_, exists := b.Step("sub+install")
		require.True(t, exists)
	})

	t.Run("SourceExists", func(t *testing.T) {
		_, exists := b.Source("sub+deps")
		require.True(t, exists)
	})
}
