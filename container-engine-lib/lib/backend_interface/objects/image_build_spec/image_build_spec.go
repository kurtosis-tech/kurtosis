package image_build_spec

type ImageBuildSpec struct {
	contextDir string
}

func NewImageBuildSpec(contextDir string) *ImageBuildSpec {
	return &ImageBuildSpec{
		contextDir: contextDir,
	}
}

func (imageBuildSpec *ImageBuildSpec) GetContextDir() string {
	return imageBuildSpec.contextDir
}
