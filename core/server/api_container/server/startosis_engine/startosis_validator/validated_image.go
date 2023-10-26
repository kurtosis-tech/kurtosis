package startosis_validator

type ValidatedImage struct {
	name string

	// whether container image was pulled from remote registry
	pulledFromRemote bool

	// whether container image was built by Kurtosis
	locallyBuilt bool
}

func NewValidatedImage(name string, pulledFromRemote bool, locallyBuilt bool) *ValidatedImage {
	return &ValidatedImage{name: name, pulledFromRemote: pulledFromRemote, locallyBuilt: locallyBuilt}
}

func (v *ValidatedImage) GetName() string {
	return v.name
}

func (v *ValidatedImage) IsPulledFromRemote() bool {
	return v.pulledFromRemote
}

func (v *ValidatedImage) IsLocallyBuilt() bool {
	return v.locallyBuilt
}
