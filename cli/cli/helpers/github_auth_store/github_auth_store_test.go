package github_auth_store

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//DO NOT CHANGE THIS VALUE
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	expectedKeyringServiceName = "kurtosis-cli"
)

// The keyring service name in this package has to be always "kurtosis-cli"
// so we control that it does not change
func TestKeyringServiceNameDoesNotChange(t *testing.T) {
	require.Equal(t, expectedKeyringServiceName, kurtosisCliKeyringServiceName)
}
