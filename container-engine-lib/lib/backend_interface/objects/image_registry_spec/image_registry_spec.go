package image_registry_spec

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
)

type ImageRegistrySpec struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateRegistrySpec *privateImageRegistrySpec
}

type privateImageRegistrySpec struct {
	Image        string
	Username     string
	Password     string
	RegistryAddr string
}

func NewImageRegistrySpec(image, username, password, registryAddr string) *ImageRegistrySpec {
	internalRegistrySpec := &privateImageRegistrySpec{
		Image:        image,
		Username:     username,
		Password:     password,
		RegistryAddr: registryAddr,
	}
	return &ImageRegistrySpec{privateRegistrySpec: internalRegistrySpec}
}

func (irs *ImageRegistrySpec) GetImageName() string {
	return irs.privateRegistrySpec.Image
}

func (irs *ImageRegistrySpec) GetUsername() string {
	return irs.privateRegistrySpec.Username
}

func (irs *ImageRegistrySpec) GetPassword() string {
	return irs.privateRegistrySpec.Password
}

func (irs *ImageRegistrySpec) GetRegistryAddr() string {
	return irs.privateRegistrySpec.RegistryAddr
}

func (irs *ImageRegistrySpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(irs.privateRegistrySpec)
}

func (irs *ImageRegistrySpec) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateImageRegistrySpec{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	irs.privateRegistrySpec = unmarshalledPrivateStructPtr
	return nil
}
