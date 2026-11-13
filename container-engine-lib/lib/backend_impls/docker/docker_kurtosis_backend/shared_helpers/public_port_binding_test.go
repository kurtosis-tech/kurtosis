package shared_helpers

import (
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
)

func getPublicBindingForHostIp(t *testing.T, hostIp string) (string, *port_spec.PortSpec) {
	privatePortSpec, err := port_spec.NewPortSpec(30303, port_spec.TransportProtocol_TCP, "", nil, "")
	require.NoError(t, err)
	bindings := map[nat.Port]*nat.PortBinding{
		"30303/tcp": {HostIP: hostIp, HostPort: "33688"},
	}
	publicIp, publicPortSpec, err := GetPublicPortBindingFromPrivatePortSpec(privatePortSpec, bindings)
	require.NoError(t, err)
	return publicIp.String(), publicPortSpec
}

func TestPublicPortBindingReportsRealBindIpAndConnectableIp(t *testing.T) {
	publicIp, publicPortSpec := getPublicBindingForHostIp(t, "0.0.0.0")
	require.Equal(t, "127.0.0.1", publicIp)
	require.Equal(t, "0.0.0.0", *publicPortSpec.GetMaybeHostIp())
	require.Equal(t, uint16(33688), publicPortSpec.GetNumber())

	publicIp, publicPortSpec = getPublicBindingForHostIp(t, "127.0.0.1")
	require.Equal(t, "127.0.0.1", publicIp)
	require.Equal(t, "127.0.0.1", *publicPortSpec.GetMaybeHostIp())
}
