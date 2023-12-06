package image_build_spec

type ImageBuildSpec struct {
	contextDir string

	containerImageFilePath string

	targetStage string
}

func NewImageBuildSpec(contextDir string, containerImageFilePath string, targetStage string) *ImageBuildSpec {
	return &ImageBuildSpec{
		contextDir:             contextDir,
		containerImageFilePath: containerImageFilePath,
		targetStage:            targetStage,
	}
}

func (imageBuildSpec *ImageBuildSpec) GetContextDir() string {
	return imageBuildSpec.contextDir
}

func (imageBuildSpec *ImageBuildSpec) GetTargetStage() string {
	return imageBuildSpec.targetStage
}

func (imageBuildSpec *ImageBuildSpec) GetContainerImageFilePath() string {
	return imageBuildSpec.containerImageFilePath
}
