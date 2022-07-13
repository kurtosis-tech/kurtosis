package kubernetes

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"testing"
)

var allPodPhases = []apiv1.PodPhase{
	apiv1.PodPending,
	apiv1.PodRunning,
	apiv1.PodSucceeded,
	apiv1.PodFailed,
	apiv1.PodUnknown,
}
func TestIsPodRunningDeterminerCompleteness(t *testing.T) {
	for _, podPhase := range allPodPhases {
		_, found := isPodRunningDeterminer[podPhase]
		require.True(t, found, "No is-running designation set for pod phase '%v'", podPhase)
	}
}

func TestKubernetesPortProtocolLookupCompleteness(t *testing.T) {
	for _, kurtosisPortProtocol := range port_spec.PortProtocolValues() {
		_, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisPortProtocol]
		require.True(t, found, "No Kubernetes port protocol defined for Kurtosis port protocol '%v'", kurtosisPortProtocol.String())
	}
}

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

	_, err = getKubernetesServicePortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
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

	_, err = getKubernetesContainerPortsFromPrivatePortSpecs(map[string]*port_spec.PortSpec{
		portId1: portSpec1,
		portId2: portSpec2,
		portId3: portSpec3,
	})
	require.NoError(t, err)
}

func TestParseCPUAllocationReturnsCorrectQuantityValue(t *testing.T){
	cpuAllocationStr := "1.5"

	cpuQuantity, err := ParseCPUAllocation(cpuAllocationStr)
	require.NoError(t, err)

	// 1.5 cpus == 1500 millicpus
	require.Equal(t, int64(1500), cpuQuantity.MilliValue()) // 1.5 cpus == 1500 millicpus
}

func TestParseCPUAllocationReturnsErrorForIncorrectFormat(t *testing.T){
	cpuAllocationStr := "1.5cpus"

	_, err := ParseCPUAllocation(cpuAllocationStr)
	require.Error(t, err)
}

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T){
	memoryAllocation := uint64(400) // 400 megabytes

	memoryAllocationInBytes := convertMemoryAllocationToBytes(memoryAllocation)
	require.Equal(t, int64(400000000), int64(memoryAllocationInBytes))
}