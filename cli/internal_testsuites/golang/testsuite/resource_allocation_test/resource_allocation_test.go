package resource_allocation_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "resource-allocation-test"
	isPartitioningEnabled = false

	resourceAllocTestImageName = "alpine:3.12.4"
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
	containerConfigSupplier := getContainerConfigSupplierWithCPUAndMemory()

	_, err = enclaveCtx.AddService(testServiceId, containerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service with the cpuAllocationMillicpus=`%d` and memoryAllocationMegabytes=`%d`", testCpuAllocMillicpus, testMemoryAllocMegabytes)
}

func TestSettingInvalidMemoryAllocationMegabytesReturnsError(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfigSupplier := getContainerConfigSupplierWithInvalidMemory()

	_, err = enclaveCtx.AddService(testServiceId, containerConfigSupplier)
	require.Error(t, err, "An error should have occurred with the following invalid memory allocation: `%d`", testInvalidMemoryAllocMegabytes)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getContainerConfigSupplierWithCPUAndMemory() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			resourceAllocTestImageName,
		).WithCPUAllocationMillicpus(
			testCpuAllocMillicpus,
		).WithMemoryAllocationMegabytes(
			testMemoryAllocMegabytes,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getContainerConfigSupplierWithInvalidMemory() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			resourceAllocTestImageName,
		).WithMemoryAllocationMegabytes(
			testInvalidMemoryAllocMegabytes,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
