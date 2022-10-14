package startosis_executor

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"

const (
	artifactUuidSuffix = "artifact_uuid"
)

type ExecutionEnvironment struct {
	filesArtifactCache map[string]string
}

func NewExecutionEnvironment() *ExecutionEnvironment {
	return &ExecutionEnvironment{
		map[string]string{},
	}
}

func (environment *ExecutionEnvironment) GetArtifactUuid(key string) (string, bool) {
	artifactUuid, found := environment.filesArtifactCache[key]
	return artifactUuid, found
}

func (environment *ExecutionEnvironment) SetArtifactUuid(position kurtosis_instruction.InstructionPosition, artifactUuid string) {
	environment.filesArtifactCache[position.MagicString(artifactUuidSuffix)] = artifactUuid
}
