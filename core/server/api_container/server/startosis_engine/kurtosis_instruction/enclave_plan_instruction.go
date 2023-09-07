package kurtosis_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/enclave_plan_capabilities"
)

type EnclavePlanInstruction interface {
	GetKurtosisInstructionStr() string
	GetCapabilities() *enclave_plan_capabilities.EnclavePlanCapabilities
	Executed(isExecuted bool)
	IsExecuted() bool
}
