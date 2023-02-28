package shared_helpers

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

func ParseSubnetworks(subnetworkArgName string, subnetworksTuple starlark.Tuple) (service_network_types.PartitionID, service_network_types.PartitionID, *startosis_errors.InterpretationError) {
	subnetworksStr, interpretationErr := kurtosis_types.SafeCastToStringSlice(subnetworksTuple, subnetworkArgName)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}
	if len(subnetworksStr) != 2 {
		return "", "", startosis_errors.NewInterpretationError("Subnetworks tuple should contain exactly 2 subnetwork names. %d was/were provided", len(subnetworksStr))
	}
	subnetwork1 := service_network_types.PartitionID(subnetworksStr[0])
	subnetwork2 := service_network_types.PartitionID(subnetworksStr[1])
	return subnetwork1, subnetwork2, nil
}
