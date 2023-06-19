package kurtosis_starlark_framework

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

type InstructionUuid string

func GenerateInstructionUuid() (InstructionUuid, error) {
	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to generate new random instruction UUID")
	}
	return InstructionUuid(uuid), nil
}
