package template

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	s := &struct {
		Hello string
	}{
		Hello: "Commit: {{ .Git.ShaShort }}, Branch: {{ .Git.Branch }}",
	}
	err := Struct(s)
	require.NoError(t, err)
	fmt.Printf("%+v\n", s)
}
