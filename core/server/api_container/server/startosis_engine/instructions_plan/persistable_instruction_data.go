package instructions_plan

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type PersistableInstructionData struct {
	kurtosisInstructionStr string
	capabilities           *PersitableCapabilities
}

type PersitableCapabilities struct {
	serviceName      service.ServiceName
	serviceNames     []service.ServiceName
	artifactName     string
	filesArtifactMD5 []byte
}

func NewPersistableInstructionData(
	kurtosisInstruction kurtosis_instruction.KurtosisInstruction,
	instructionCapabilities kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities,
) *PersistableInstructionData {
	persistableInstructionData := &PersistableInstructionData{
		kurtosisInstructionStr: kurtosisInstruction.String(),
	}
	return persistableInstructionData
}

func (instructionData *PersistableInstructionData) GetKurtosisInstructionStr() string {
	return instructionData.kurtosisInstructionStr
}

func createPersistableCapabilitiesFromInstructionCapabilities(instructionCapabilities kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities) (*PersitableCapabilities, error){

	newPersistableCapabitilies := &PersitableCapabilities{}

	switch value := instructionCapabilities.(type) {
	case add_service.AddServiceCapabilities:
		addServiceCapabilites, ok := instructionCapabilities.(*add_service.AddServiceCapabilities)
		if !ok {
			return nil, stacktrace.NewError("An error occurred when casting to AddServiceCapabilities") //TODO improve
		}
		newPersistableCapabitilies.serviceName = addServiceCapabilites.
		return value, nil
	case *starlark.List, *starlark.Set, starlark.Tuple:
		return value, nil
	}
}
