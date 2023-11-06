package metrics_client

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const envVarValueForTesting = "testing"

func TestISCIWhenEnvironmentVariableIsSetSucceeds(t *testing.T) {
	for _, envVar := range ciEnvironmentVariables {
		err := os.Setenv(envVar, envVarValueForTesting)
		defer os.Unsetenv(envVar)
		require.Nil(t, err)
		require.True(t, IsCI())
	}
}
