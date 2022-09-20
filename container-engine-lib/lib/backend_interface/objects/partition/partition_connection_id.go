package partition

import "strings"

/*
Represents two partitions, where order is unimportant
*/
type PartitionConnectionID struct {
	lexicalFirst  PartitionID
	lexicalSecond PartitionID
}

// NOTE: It's very important that the constructor is used here!
func NewPartitionConnectionID(partitionA PartitionID, partitionB PartitionID) *PartitionConnectionID {

	// We sort these upon creation so that this type can be used as a key in a map, and so that
	// 	this tuple is commutative: partitionConnectionID(A, B) == partitionConnectionID(B, A) as a map key
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

func (id PartitionConnectionID) GetFirst() PartitionID {
	return id.lexicalFirst
}

func (id PartitionConnectionID) GetSecond() PartitionID {
	return id.lexicalSecond
}
