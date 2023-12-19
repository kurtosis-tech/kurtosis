package startosis_validator

type ValidatedImage struct {
	name             string
	pulledFromRemote bool
	builtLocally     bool
	architecture     string
}

func NewValidatedImage(name string, pulledFromRemote bool, builtLocally bool, architecture string) *ValidatedImage {
	return &ValidatedImage{
		name:             name,
		pulledFromRemote: pulledFromRemote,
		builtLocally:     builtLocally,
		architecture:     architecture}
}

func (v *ValidatedImage) GetName() string {
	return v.name
}

func (v *ValidatedImage) IsPulledFromRemote() bool {
	return v.pulledFromRemote
}

func (v *ValidatedImage) IsBuiltLocally() bool {
	return v.builtLocally
}

func (v *ValidatedImage) GetArchitecture() string {
	return v.architecture
}
