package packet_delay_distribution

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
)

type PacketDelayDistribution interface {
	builtin_argument.KurtosisValueType

	ToKurtosisType() (*partition_topology.PacketDelayDistribution, *startosis_errors.InterpretationError)
}
