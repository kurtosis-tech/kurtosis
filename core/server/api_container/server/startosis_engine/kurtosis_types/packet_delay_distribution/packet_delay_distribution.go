package packet_delay_distribution

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
)

type PacketDelayDistribution interface {
	kurtosis_type_constructor.KurtosisValueType

	ToKurtosisType() (*partition_topology.PacketDelayDistribution, *startosis_errors.InterpretationError)
}
