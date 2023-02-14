package kurtosis_config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// This test allows us to verify that a v0 config can go all the way up to latest
// I.e., that people don't forget to add config-override schema migrations for the new versions they add
func TestMigrateConfigOverridesToLatest(t *testing.T) {
	v0ConfigFileBytes := []byte(`{"shouldSendMetrics":true}`)
	_, err := migrateConfigOverridesToLatest(v0ConfigFileBytes)
	require.NoError(t, err)
}
