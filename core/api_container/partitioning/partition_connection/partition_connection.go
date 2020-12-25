/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partition_connection

import (
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning"
	"strings"
)

type PartitionConnection struct {
	IsBlocked bool
}

/*
Represents two partitions, where order is unimportant
 */
type PartitionConnectionID struct {
	lexicalFirst  partitioning.PartitionID
	lexicalSecond partitioning.PartitionID
}

func NewPartitionConnectionID(partitionA partitioning.PartitionID, partitionB partitioning.PartitionID) *PartitionConnectionID {

	// We sort these upon creation so that this type can be used as a key in a map, and so that
	// 	this tuple is commutative: PartitionConnectionID(A, B) == PartitionConnectionID(B, A) as a map key
	first, second := partitionA, partitionB
	result := strings.Compare(string(first), string(second))
	if result > 0 {
		first, second = second, first
	}
	return &PartitionConnectionID{
		lexicalFirst:  first,
		lexicalSecond: second,
	}
}
