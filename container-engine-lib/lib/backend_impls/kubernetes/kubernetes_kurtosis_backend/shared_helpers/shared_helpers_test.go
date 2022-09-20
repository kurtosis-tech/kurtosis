package shared_helpers

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetServicePortsFromPortSpecs(t *testing.T) {
	portId1 := "tcp"
	portId2 := "udp"
	portId3 := "sctp"

	portSpec1, err := port_spec.NewPortSpec(100, port_spec.PortProtocol_TCP)
	require.NoError(t, err)
	portSpec2, err := port_spec.NewPortSpec(200, port_spec.PortProtocol_UDP)
	require.NoError(t, err)
	portSpec3, err := port_spec.NewPortSpec(300, port_spec.PortProtocol_SCTP)
	require.NoError(t, err)

	_, err = GetKubernetesServicePortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		portId1: portSpec1,
		portId2: portSpec2,
		portId3: portSpec3,
	})
	require.NoError(t, err)
}

func TestGetContainerPortsFromPortSpecs(t *testing.T) {
	portId1 := "tcp"
	portId2 := "udp"
	portId3 := "sctp"

	portSpec1, err := port_spec.NewPortSpec(100, port_spec.PortProtocol_TCP)
	require.NoError(t, err)
	portSpec2, err := port_spec.NewPortSpec(200, port_spec.PortProtocol_UDP)
	require.NoError(t, err)
	portSpec3, err := port_spec.NewPortSpec(300, port_spec.PortProtocol_SCTP)
	require.NoError(t, err)

	_, err = GetKubernetesContainerPortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		portId1: portSpec1,
		portId2: portSpec2,
		portId3: portSpec3,
	})
	require.NoError(t, err)
}