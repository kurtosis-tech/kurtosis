/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_id_set

import "github.com/kurtosis-tech/kurtosis/api_container/partitioning"

// Stupid Go... why can't it just have generics so all this is part of stdlib
type ServiceIDSet struct {
	elems map[partitioning.ServiceID]bool
}

func NewServiceIDSet() *ServiceIDSet {
	return &ServiceIDSet{
		elems: map[partitioning.ServiceID]bool{},
	}
}

func (set *ServiceIDSet) Add(elem partitioning.ServiceID) {
	set.elems[elem] = true
}

func (set ServiceIDSet) Copy() ServiceIDSet {
	elemsCopy := map[partitioning.ServiceID]bool{}
	for elem, _ := range set.elems {
		elemsCopy[elem] = true
	}
	return ServiceIDSet{elems: elemsCopy}
}

func (set ServiceIDSet) Equal(other ServiceIDSet) bool {
	if set.Size() != other.Size() {
		return false
	}
	for elem, _ := range set.elems {
		if !other.Contains(elem) {
			return false
		}
	}
	return true

}

func (set ServiceIDSet) Size() int {
	return len(set.elems)
}

func (set ServiceIDSet) Contains(elem partitioning.ServiceID) bool {
	_, found := set.elems[elem]
	return found
}

func (set ServiceIDSet) Elems() []partitioning.ServiceID {
	result := []partitioning.ServiceID{}
	for elem, _ := range set.elems {
		result = append(result, elem)
	}
	return result
}
