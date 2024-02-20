package image_spec

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
)

type ImageSpec struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateRegistrySpec *privateImageSSpec
}

type privateImageSSpec struct {
	Image        string
	Username     string
	Password     string
	RegistryAddr string
}

func NewImagSpec(image, username, password, registryAddr string) *ImageSpec {
	internalRegistrySpec := &privateImageSSpec{
		Image:        image,
		Username:     username,
		Password:     password,
		RegistryAddr: registryAddr,
	}
	return &ImageSpec{privateRegistrySpec: internalRegistrySpec}
}

func (irs *ImageSpec) GetImageName() string {
	return irs.privateRegistrySpec.Image
}

func (irs *ImageSpec) GetUsername() string {
	return irs.privateRegistrySpec.Username
}

func (irs *ImageSpec) GetPassword() string {
	return irs.privateRegistrySpec.Password
}

func (irs *ImageSpec) GetRegistryAddr() string {
	return irs.privateRegistrySpec.RegistryAddr
}

func (irs *ImageSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(irs.privateRegistrySpec)
}

func (irs *ImageSpec) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateImageSSpec{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	irs.privateRegistrySpec = unmarshalledPrivateStructPtr
	return nil
}
