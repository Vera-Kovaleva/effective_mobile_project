package pointer_test

import (
	"testing"
	"time"

	"ef_project/internal/infra/pointer"

	"github.com/stretchr/testify/require"
)

func TestRef(t *testing.T) {
	t.Parallel()

	value := "some value"

	require.Equal(t, value, *pointer.Ref(value))
}

func TestDates(t *testing.T) {
	input := "2025-07"

	value, err := time.Parse("2006-01", input)
	require.NoError(t, err)
	_ = value

	output := value.Format("2006_01")
	require.Equal(t, "2025_07", output)
}
