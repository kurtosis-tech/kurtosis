package startosis_executor

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

func (environment *ExecutionEnvironment) SetArtifactUuid(key, artifactUuid string) {
	environment.filesArtifactCache[key] = artifactUuid
}
