package image_registry_spec

type ImageRegistrySpec struct {
	image        string
	email        string
	username     string
	password     string
	registryAddr string
}

func NewImageRegistrySpec(image, email, username, password, registryAddr string) *ImageRegistrySpec {
	return &ImageRegistrySpec{
		image:        image,
		email:        email,
		username:     username,
		password:     password,
		registryAddr: registryAddr,
	}
}

func (irs *ImageRegistrySpec) GetImage() string {
	return irs.image
}

func (irs *ImageRegistrySpec) GetEmail() string {
	return irs.email
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
