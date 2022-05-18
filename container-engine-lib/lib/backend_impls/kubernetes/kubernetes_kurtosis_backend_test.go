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