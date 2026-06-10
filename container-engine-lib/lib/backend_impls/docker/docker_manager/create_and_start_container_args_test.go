package docker_manager

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAndStartContainerArgs_WithPrivileged(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("img", "name", "net").
		WithPrivileged(true).
		Build()
	require.True(t, args.privileged)
}

func TestCreateAndStartContainerArgs_PrivilegedDefaultsToFalse(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("img", "name", "net").Build()
	require.False(t, args.privileged)
}

func TestCreateAndStartContainerArgs_BindMountsRoundTrip(t *testing.T) {
	bindMounts := map[string]string{"/var/run/docker.sock": "/var/run/docker.sock"}
	args := NewCreateAndStartContainerArgsBuilder("img", "name", "net").
		WithBindMounts(bindMounts).
		Build()
	require.Equal(t, bindMounts, args.bindMounts)
}

func TestCreateAndStartContainerArgs_WithHostPIDNamespace(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("img", "name", "net").
		WithHostPIDNamespace(true).
		Build()
	require.True(t, args.hostPIDNamespace)
}

func TestCreateAndStartContainerArgs_HostPIDNamespaceDefaultsToFalse(t *testing.T) {
	args := NewCreateAndStartContainerArgsBuilder("img", "name", "net").Build()
	require.False(t, args.hostPIDNamespace)
}
