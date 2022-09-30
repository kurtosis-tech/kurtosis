package resource_allocation_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "resource-allocation-test"
	isPartitioningEnabled = false

	resourceAllocTestImageName = "flashspys/nginx-static"
	testServiceId              = "test"

	testMemoryAllocMegabytes        = 1000 // 10000 megabytes = 1 GB
	testCpuAllocMillicpus           = 1000 // 1000 millicpus = 1 CPU
	testInvalidMemoryAllocMegabytes = 4    // 6 megabytes is Dockers min, so this should throw error
)

func TestSettingResourceAllocationFieldsAddsServiceWithNoError(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfig := getContainerConfigWithCPUAndMemory()

	_, err = enclaveCtx.AddService(testServiceId, containerConfig)
	require.NoError(t, err, "An error occurred adding the file server service with the cpuAllocationMillicpus=`%d` and memoryAllocationMegabytes=`%d`", testCpuAllocMillicpus, testMemoryAllocMegabytes)
}

func TestSettingInvalidMemoryAllocationMegabytesReturnsError(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfig := getContainerConfigWithInvalidMemory()

	_, err = enclaveCtx.AddService(testServiceId, containerConfig)
	require.Error(t, err, "An error should have occurred with the following invalid memory allocation: `%d`", testInvalidMemoryAllocMegabytes)
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func getContainerConfigWithCPUAndMemory() *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(
		resourceAllocTestImageName,
	).WithCPUAllocationMillicpus(
		testCpuAllocMillicpus,
	).WithMemoryAllocationMegabytes(
		testMemoryAllocMegabytes,
	).Build()
	return containerConfig
}

func getContainerConfigWithInvalidMemory() *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(
		resourceAllocTestImageName,
	).WithMemoryAllocationMegabytes(
		testInvalidMemoryAllocMegabytes,
	).Build()
	return containerConfig
}
