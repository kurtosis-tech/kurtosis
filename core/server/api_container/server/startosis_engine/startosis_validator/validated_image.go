package startosis_validator

type ValidatedImage struct {
	name             string
	pulledFromRemote bool
	architecture     string
}

func NewValidatedImage(name string, pulledFromRemote bool, architecture string) *ValidatedImage {
	return &ValidatedImage{name: name, pulledFromRemote: pulledFromRemote, architecture: architecture}
}

func (v *ValidatedImage) GetName() string {
	return v.name
}

func (v *ValidatedImage) GetPulledFromRemote() bool {
	return v.pulledFromRemote
}

func (v *ValidatedImage) GetArchitecture() string {
	return v.architecture
}
