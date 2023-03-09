package service_partitions

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceA   = service.ServiceName("test-service-a")
	testServiceB   = service.ServiceName("test-service-b")
	invalidService = service.ServiceName("invalid-service")

	testPartitionA = partition.PartitionID("test-partition-a")
	testPartitionB = partition.PartitionID("test-partition-b")
)

func TestGetService_ForExistingServiceSucceeds(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	servicePartitions, err := GetOrCreateServicePartitionsBucket(enclaveDb)
	require.Nil(t, err)
	err = servicePartitions.AddPartitionToService(testServiceA, testPartitionA)
	require.Nil(t, err)
	exists, err := servicePartitions.DoesServiceExist(testServiceA)
	require.Nil(t, err)
	require.True(t, exists)
	partitionForServiceResult, err := servicePartitions.GetPartitionForService(testServiceA)
	require.Nil(t, err)
	require.Equal(t, testPartitionA, partitionForServiceResult)
}

func TestGetService_FailsForInvalidService(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	servicePartitions, err := GetOrCreateServicePartitionsBucket(enclaveDb)
	require.Nil(t, err)
	exists, err := servicePartitions.DoesServiceExist(invalidService)
	require.Nil(t, err)
	require.False(t, exists)
	partitionForServiceResult, err := servicePartitions.GetPartitionForService(invalidService)
	require.Empty(t, partitionForServiceResult)
	require.Nil(t, err)
}

func TestReplaceBucketContents(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	servicePartitions, err := GetOrCreateServicePartitionsBucket(enclaveDb)
	require.Nil(t, err)
	err = servicePartitions.AddPartitionToService(testServiceA, testPartitionA)
	require.Nil(t, err)

	newServicePartitionMap := map[service.ServiceName]partition.PartitionID{
		testServiceB: testPartitionB,
	}

	err = servicePartitions.ReplaceBucketContents(newServicePartitionMap)
	require.Nil(t, err)
	actualNewMapping, err := servicePartitions.GetAllServicePartitions()
	require.Nil(t, err)
	require.Equal(t, newServicePartitionMap, actualNewMapping)
}

func TestRemoveService_ForExistingServiceSucceeds(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	servicePartitions, err := GetOrCreateServicePartitionsBucket(enclaveDb)
	require.Nil(t, err)
	err = servicePartitions.AddPartitionToService(testServiceA, testPartitionA)
	require.Nil(t, err)
	exists, err := servicePartitions.DoesServiceExist(testServiceA)
	require.Nil(t, err)
	require.True(t, exists)
	err = servicePartitions.RemoveService(testServiceA)
	require.Nil(t, err)
	exists, err = servicePartitions.DoesServiceExist(testServiceA)
	require.Nil(t, err)
	require.False(t, exists)
}
