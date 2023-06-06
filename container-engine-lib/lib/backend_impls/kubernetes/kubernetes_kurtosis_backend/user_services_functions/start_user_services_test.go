package user_services_functions

import (
	"fmt"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"testing"
)

func TestConvertMemoryAllocationToBytesReturnsCorrectValue(t *testing.T) {
	memoryAllocationMegabytes := uint64(400) // 400 megabytes

	memoryAllocationBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
	require.Equal(t, uint64(400000000), memoryAllocationBytes)
}

func Test_checkIfResourcesAreSetProperly(t *testing.T) {
	resourceRequst := uint64(100)
	resourceLimit := uint64(99)

	err := checkIfResourcesAreSetProperly(resourceRequst, resourceLimit, apiv1.ResourceCPU)
	require.NotNil(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("Minimum Resource Requirement for the container is set higher than Maximum resource requirement for resource: %v", apiv1.ResourceCPU))
}
