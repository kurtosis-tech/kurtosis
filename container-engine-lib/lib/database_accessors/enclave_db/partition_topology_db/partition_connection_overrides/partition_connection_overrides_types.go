package partition_connection_overrides

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"

// PartitionConnectionID The fields have to be upper-cased for JSON serialization to work
type PartitionConnectionID struct {
	LexicalFirst  partition.PartitionID `json:"lexical_first"`
	LexicalSecond partition.PartitionID `json:"lexical_second"`
}

// DelayDistribution The fields have to be upper cased for JSON serialization to work
type DelayDistribution struct {
	AvgDelayMs  uint32  `json:"avg_delay"`
	Jitter      uint32  `json:"jitter"`
	Correlation float32 `json:"correlation"`
}

var EmptyPartitionConnection PartitionConnection

// PartitionConnection The fields have to be upper-cased for JSON serialization to work
type PartitionConnection struct {
	PacketLoss              float32           `json:"packet_loss"`
	PacketDelayDistribution DelayDistribution `json:"delay_distribution"`
}
