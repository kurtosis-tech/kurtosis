package image_registry_spec

type ImageRegistrySpec struct {
	image        string
	username     string
	password     string
	registryAddr string
}

func NewImageRegistrySpec(image, username, password, registryAddr string) *ImageRegistrySpec {
	return &ImageRegistrySpec{
		image:        image,
		username:     username,
		password:     password,
		registryAddr: registryAddr,
	}
}

func (irs *ImageRegistrySpec) GetImage() string {
	return irs.image
}

func (irs *ImageRegistrySpec) GetUsername() string {
	return irs.username
}

func (irs *ImageRegistrySpec) GetPassword() string {
	return irs.password
}

func (irs *ImageRegistrySpec) GetRegistryAddr() string {
	return irs.registryAddr
}
