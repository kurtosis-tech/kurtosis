package startosis_executor

type ExecutionEnvironment struct {
	filesArtifactMap map[string]string
}

func NewExecutionEnvironment() *ExecutionEnvironment {
	return &ExecutionEnvironment{
		map[string]string{},
	}
}

// GetArtifactUuid is used during the execution time to get the artifact uuid of an instruction
// previously assigned by SetArtifactUuid. Users should use this in the KurtosisInstruction.Execute phase.
func (environment *ExecutionEnvironment) GetArtifactUuid(artifactUuidMagicString string) (string, bool) {
	artifactUuid, found := environment.filesArtifactMap[artifactUuidMagicString]
	return artifactUuid, found
}

// SetArtifactUuid is used during execution time to add real artifact uuid for magic strings created
// during interpretation. Users should use this in the KurtosisInstruction.Execute phase
func (environment *ExecutionEnvironment) SetArtifactUuid(artifactUuidMagicString, artifactUuid string) {
	environment.filesArtifactMap[artifactUuidMagicString] = artifactUuid
}
