package partition_connection_overrides

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"

type PartitionConnectionID struct {
	LexicalFirst  partition.PartitionID `json:"lexical_first"`
	LexicalSecond partition.PartitionID `json:"lexical_second"`
}

type DelayDistribution struct {
	AvgDelayMs  uint32  `json:"avg_delay"`
	Jitter      uint32  `json:"jitter"`
	Correlation float32 `json:"correlation"`
}

var EmptyPartitionConnection PartitionConnection

type PartitionConnection struct {
	PacketLoss              float32           `json:"packet_loss"`
	PacketDelayDistribution DelayDistribution `json:"delay_distribution"`
}
