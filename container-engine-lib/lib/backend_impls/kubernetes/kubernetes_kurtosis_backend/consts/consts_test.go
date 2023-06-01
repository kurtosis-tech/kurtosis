package consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"testing"
)

var allPodPhases = []v1.PodPhase{
	v1.PodPending,
	v1.PodRunning,
	v1.PodSucceeded,
	v1.PodFailed,
	v1.PodUnknown,
}

func TestIsPodRunningDeterminerCompleteness(t *testing.T) {
	for _, podPhase := range allPodPhases {
		_, found := IsPodRunningDeterminer[podPhase]
		require.True(t, found, "No is-running designation set for pod phase '%v'", podPhase)
	}
}

func TestKubernetesTransportProtocolLookupCompleteness(t *testing.T) {
	for _, kurtosisTransportProtocol := range port_spec.TransportProtocolValues() {
		_, found := KurtosisTransportProtocolToKubernetesTransportProtocolTranslator[kurtosisTransportProtocol]
		require.True(t, found, "No Kubernetes port protocol defined for Kurtosis port protocol '%v'", kurtosisTransportProtocol.String())
	}
}
