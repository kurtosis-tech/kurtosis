package kurtosis_types

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"

type PacketDelayDistributionInterface interface {
	ToKurtosisType() partition_topology.PacketDelayDistribution
}
