package starlark_warning

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStarlarkWarningMessage(t *testing.T) {
	// add warnings
	PrintOnceAtTheEndOfExecutionf("warning one")
	PrintOnceAtTheEndOfExecutionf("warning %v", "second")

	warnings := GetContentFromWarningSet()
	require.Equal(t, 2, len(warnings))
	require.Equal(t, []string{"warning one", "warning second"}, warnings)

	// this ensures that `GetContentFromWarningSet` actually deletes all the previous warnings
	PrintOnceAtTheEndOfExecutionf("warning three")
	warnings = GetContentFromWarningSet()
	require.Equal(t, 1, len(warnings))
	require.Equal(t, []string{"warning three"}, warnings)
}
