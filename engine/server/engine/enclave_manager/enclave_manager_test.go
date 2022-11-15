package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOneToOneMappingBetweenObjAttrProtosAndDockerProtos(t *testing.T) {
	// Ensure all obj attr protos are used
	require.Equal(t, len(schema.AllowedProtocols), len(objAttrsSchemaPortProtosToDockerPortProtos))

	// Ensure all teh declared obj attr protos are valid
	for candidateObjAttrProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		_, found := schema.AllowedProtocols[candidateObjAttrProto]
		require.True(t, found, "Invalid object attribute schema proto '%v'", candidateObjAttrProto)
	}

	// Ensure no duplicate Docker protos, which is the best we can do since Docker doesn't expose an enum of all the protos they support
	seenDockerProtos := map[string]schema.PortProtocol{}
	for objAttrProto, dockerProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		preexistingObjAttrProto, found := seenDockerProtos[dockerProto]
		require.False(t, found, "Docker proto '%v' is already in use by obj attr proto '%v'", dockerProto, preexistingObjAttrProto)
		seenDockerProtos[dockerProto] = objAttrProto
	}
}

func TestIsContainerRunningDeterminerCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, found := isContainerRunningDeterminer[containerStatus]
		require.True(t, found, "No is-container-running determination provided for container status '%v'", containerStatus.String())
	}
}

func TestGetEnclaveContainersStatusFromEnclaveStatusCompleteness(t *testing.T) {
	for _, enclaveStatus := range enclave.EnclaveStatusValues() {
		_, err := getEnclaveContainersStatusFromEnclaveStatus(enclaveStatus)
		require.NoError(t, err, "No EnclaveContainersStatus provided for enclave status '%v'", enclaveStatus.String())
	}
}

func TestGetApiContainerStatusFromContainerStatusCompleteness(t *testing.T) {
	for _, containerStatus := range container_status.ContainerStatusValues() {
		_, err := getApiContainerStatusFromContainerStatus(containerStatus)
		require.NoError(t, err, "No ApiContainerStatus provided for container status '%v'", containerStatus.String())
	}
}
