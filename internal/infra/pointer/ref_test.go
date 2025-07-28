package pointer_test

import (
	"testing"

	"ef_project/internal/infra/pointer"

	"github.com/stretchr/testify/require"
)

func TestRef(t *testing.T) {
	t.Parallel()

	value := "some value"

	require.Equal(t, value, *pointer.Ref(value))
}
