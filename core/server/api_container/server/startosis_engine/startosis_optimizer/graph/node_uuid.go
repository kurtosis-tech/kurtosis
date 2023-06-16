package graph

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

type NodeUuid string

func GenerateNodeUuid() (NodeUuid, error) {
	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to generate Node UUID")
	}
	return NodeUuid(uuid), nil
}
