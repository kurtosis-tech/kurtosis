package startosis_validator

type ValidatedImage struct {
	name             string
	pulledFromRemote bool
}

func NewValidatedImage(name string, pulledFromRemote bool) *ValidatedImage {
	return &ValidatedImage{name: name, pulledFromRemote: pulledFromRemote}
}

func (v *ValidatedImage) GetName() string {
	return v.name
}

func (v *ValidatedImage) GetPulledFromRemote() bool {
	return v.pulledFromRemote
}
