package image_build_spec

// ImageBuildSpec contains the information need for building a container image.
type ImageBuildSpec struct {
	// Location of the container image to build (eg. Dockerfile) on the machine
	containerImageFilePath string

	// Location of the build context needed for building the container image
	// Build context are the files, directories, config referenced in the container image used to build the image
	contextDirPath string

	// For multi-stage image builds, targetStage specifies the stage to actually build
	// For Docker: using the same Dockerfile, you can specify building multiple binaries in the same Dockerfile.
	// Then, you can specify the targetStage to instruct Docker which binary to build when building the image.
	// eg. (info here: https://docs.docker.com/build/guide/multi-stage/#build-targets)
	// ...
	//+ FROM scratch AS client
	//+ COPY --from=build-client /bin/client /bin/
	//+ ENTRYPOINT [ "/bin/client" ]
	//
	//+ FROM scratch AS server
	//+ COPY --from=build-server /bin/server /bin/
	//+ ENTRYPOINT [ "/bin/server" ]
	//  ...
	// targetStage could be set to "server" or "client",
	// Default value is the empty string if the image build is not multi-stage.
	//
	targetStage string
}

func NewImageBuildSpec(contextDirPath string, containerImageFilePath string, targetStage string) *ImageBuildSpec {
	return &ImageBuildSpec{
		containerImageFilePath: containerImageFilePath,
		contextDirPath:         contextDirPath,
		targetStage:            targetStage,
	}
}

func (imageBuildSpec *ImageBuildSpec) GetContainerImageFilePath() string {
	return imageBuildSpec.containerImageFilePath
}

func (imageBuildSpec *ImageBuildSpec) GetContextDirPath() string {
	return imageBuildSpec.contextDirPath
}

func (imageBuildSpec *ImageBuildSpec) GetTargetStage() string {
	return imageBuildSpec.targetStage
}
