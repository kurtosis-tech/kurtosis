package startosis_validator

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	requiredDockerImages map[string]bool
}

func NewValidatorEnvironment() *ValidatorEnvironment {
	return &ValidatorEnvironment{
		map[string]bool{},
	}
}

func (environment *ValidatorEnvironment) AppendRequiredDockerImage(dockerImage string) {
	environment.requiredDockerImages[dockerImage] = true
}
