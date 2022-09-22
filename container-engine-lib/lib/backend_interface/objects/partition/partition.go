package partition

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"

type PartitionID string

// Object that represents POINT-IN-TIME information about a network partition
// Store this object and continue to reference it at your own risk!!!
type Partition struct {
	id PartitionID
	services []*service.Service
}

func NewPartition(id PartitionID, services []*service.Service) *Partition {
	//TODO add no duplicated service validation
	return &Partition{id: id, services: services}
}

func (partition *Partition) GetID() PartitionID {
	return partition.id
}

func (partition *Partition) GetServices() []*service.Service {
	return partition.services
}
