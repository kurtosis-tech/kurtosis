package startosis_executor

type ExecutionEnvironment struct {
	filesArtifactCache map[string]string
}

func NewExecutionEnvironment() *ExecutionEnvironment {
	return &ExecutionEnvironment{
		map[string]string{},
	}
}

// GetArtifactUuid is used during the execution time to get the artifact uuid of an instruction
// previously assigned by SetArtifactUuid. Users should use this in the KurtosisInstruction.Execute phase.
func (environment *ExecutionEnvironment) GetArtifactUuid(key string) (string, bool) {
	artifactUuid, found := environment.filesArtifactCache[key]
	return artifactUuid, found
}

// SetArtifactUuid is used during execution time to add real artifact uuid for magic strings created
// during interpretation. Users should use this in the KurtosisInstruction.Execute phase
func (environment *ExecutionEnvironment) SetArtifactUuid(key, artifactUuid string) {
	environment.filesArtifactCache[key] = artifactUuid
}
