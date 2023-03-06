package partition_topology_db

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	testConnectionIdA = PartitionConnectionID{
		LexicalFirst:  "apple",
		LexicalSecond: "bat",
	}

	testConnectionA = PartitionConnection{
		PacketLoss: 93,
		PacketDelayDistribution: DelayDistribution{
			AvgDelayMs:  5,
			Jitter:      10,
			Correlation: 32,
		},
	}

	testConnectionIdB = PartitionConnectionID{
		LexicalFirst:  "cat",
		LexicalSecond: "dog",
	}

	testConnectionB = PartitionConnection{
		PacketLoss: 96,
		PacketDelayDistribution: DelayDistribution{
			AvgDelayMs:  6,
			Jitter:      11,
			Correlation: 33,
		},
	}
)

func TestPartitionConnection_AddAndGetAll(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	partitionConnections, err := GetOrCreatePartitionConnectionBucket(enclaveDb)
	require.Nil(t, err)

	err = partitionConnections.AddPartitionConnection(testConnectionIdA, testConnectionA)
	require.Nil(t, err)

	allConnections, err := partitionConnections.GetAllPartitionConnections()
	require.Nil(t, err)
	expectedConnections := map[PartitionConnectionID]PartitionConnection{testConnectionIdA: testConnectionA}
	require.Equal(t, len(expectedConnections), len(allConnections))
	require.Equal(t, expectedConnections, allConnections)
}

func TestPartitionConnection_ReplaceBucketContents(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	partitionConnections, err := GetOrCreatePartitionConnectionBucket(enclaveDb)
	require.Nil(t, err)

	err = partitionConnections.AddPartitionConnection(testConnectionIdA, testConnectionA)
	require.Nil(t, err)

	replacedConnections := map[PartitionConnectionID]PartitionConnection{testConnectionIdB: testConnectionB}
	err = partitionConnections.ReplaceBucketContents(replacedConnections)
	require.Nil(t, err)

	allConnections, err := partitionConnections.GetAllPartitionConnections()
	require.Nil(t, err)
	require.Equal(t, replacedConnections, allConnections)
}

func TestPartitionConnection_GetPartitionConnection(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	partitionConnections, err := GetOrCreatePartitionConnectionBucket(enclaveDb)
	require.Nil(t, err)

	err = partitionConnections.AddPartitionConnection(testConnectionIdA, testConnectionA)
	require.Nil(t, err)

	connection, err := partitionConnections.GetPartitionConnection(testConnectionIdA)
	require.Nil(t, err)
	require.Equal(t, testConnectionA, connection)
}

func TestPartitionConnection_DeleteConnection(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	partitionConnections, err := GetOrCreatePartitionConnectionBucket(enclaveDb)
	require.Nil(t, err)

	err = partitionConnections.AddPartitionConnection(testConnectionIdA, testConnectionA)
	require.Nil(t, err)

	err = partitionConnections.RemovePartitionConnection(testConnectionIdA)
	require.Nil(t, err)

	allConnections, err := partitionConnections.GetAllPartitionConnections()
	require.Nil(t, err)
	require.Empty(t, allConnections)
}
