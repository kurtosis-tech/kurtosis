/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

type PartitionID string

type ServiceID string

type PartitionTopology struct {
	defaultConnection PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionToServices map[PartitionID]map[ServiceID]bool  // partitionId -> set<serviceId>
	servicesToPartitions map[ServiceID]PartitionID
}

type ipTablesChangesNeeded struct {
	// TODO
}
