package user_service_functions

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArePrivilegedContainersAllowed_UnsetDefaultsToTrue(t *testing.T) {
	previous, hadPrevious := os.LookupEnv(allowPrivilegedContainersEnvVar)
	t.Cleanup(func() {
		if hadPrevious {
			require.NoError(t, os.Setenv(allowPrivilegedContainersEnvVar, previous))
		} else {
			require.NoError(t, os.Unsetenv(allowPrivilegedContainersEnvVar))
		}
	})
	require.NoError(t, os.Unsetenv(allowPrivilegedContainersEnvVar))
	require.True(t, arePrivilegedContainersAllowed())
}

func TestArePrivilegedContainersAllowed_FalseDisables(t *testing.T) {
	t.Setenv(allowPrivilegedContainersEnvVar, "false")
	require.False(t, arePrivilegedContainersAllowed())
}

func TestArePrivilegedContainersAllowed_FalseIsCaseInsensitive(t *testing.T) {
	t.Setenv(allowPrivilegedContainersEnvVar, "FALSE")
	require.False(t, arePrivilegedContainersAllowed())
}

func TestArePrivilegedContainersAllowed_FalseTolerantOfWhitespace(t *testing.T) {
	t.Setenv(allowPrivilegedContainersEnvVar, "  false  ")
	require.False(t, arePrivilegedContainersAllowed())
}

func TestArePrivilegedContainersAllowed_TrueValueIsAllowed(t *testing.T) {
	t.Setenv(allowPrivilegedContainersEnvVar, "true")
	require.True(t, arePrivilegedContainersAllowed())
}

func TestArePrivilegedContainersAllowed_OtherValuesAllowed(t *testing.T) {
	// Anything that isn't "false" leaves privileged enabled — fail-open.
	t.Setenv(allowPrivilegedContainersEnvVar, "anything")
	require.True(t, arePrivilegedContainersAllowed())
}
