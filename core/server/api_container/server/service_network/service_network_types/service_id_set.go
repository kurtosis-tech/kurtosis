/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network_types

/*
// Stupid Go... why can't it just have generics so all this is part of stdlib
type ServiceIDSet struct {
	elems map[ServiceID]bool
}

func NewServiceIDSet(services ...ServiceID) *ServiceIDSet {
	elems := map[ServiceID]bool{}
	for _, service := range services {
		elems[service] = true
	}
	return &ServiceIDSet{
		elems: elems,
	}
}

func (set *ServiceIDSet) AddElem(elem ServiceID) {
	set.elems[elem] = true
}

func (set *ServiceIDSet) AddElems(elems *ServiceIDSet) {
	for _, elem := range elems.Elems() {
		set.AddElem(elem)
	}
}

func (set ServiceIDSet) Copy() *ServiceIDSet {
	return NewServiceIDSet(set.Elems()...)
}

func (set ServiceIDSet) Equals(other *ServiceIDSet) bool {
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

func (set ServiceIDSet) Contains(elem ServiceID) bool {
	_, found := set.elems[elem]
	return found
}

func (set ServiceIDSet) Elems() []ServiceID {
	result := []ServiceID{}
	for elem, _ := range set.elems {
		result = append(result, elem)
	}
	return result
}

func (set *ServiceIDSet) RemoveElem(elem ServiceID) {
	delete(set.elems, elem)
}

// Removes the given elems from the set, if they exist
func (set *ServiceIDSet) RemoveElems(elems *ServiceIDSet) {
	for _, elem := range elems.Elems() {
		set.RemoveElem(elem)
	}
}

*/
