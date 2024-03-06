package image_build_spec

import (
	"encoding/json"

	"github.com/kurtosis-tech/stacktrace"
)

type ImageBuildSpec struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateImageBuildSpec *privateImageBuildSpec
}

// ImageBuildSpec contains the information need for building a container image.
type privateImageBuildSpec struct {
	// Location of the container image to build (eg. Dockerfile) on the machine
	ContainerImageFilePath string

	// Location of the build context needed for building the container image
	// Build context are the files, directories, config referenced in the container image used to build the image
	ContextDirPath string

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
	TargetStage string

	// Dockerfile build args
	BuildArgs map[string]*string
}

func NewImageBuildSpec(contextDirPath string, containerImageFilePath string, targetStage string, buildArgs map[string]*string) *ImageBuildSpec {
	internalImageBuildSpec := &privateImageBuildSpec{
		ContainerImageFilePath: containerImageFilePath,
		ContextDirPath:         contextDirPath,
		TargetStage:            targetStage,
		BuildArgs:              buildArgs,
	}
	return &ImageBuildSpec{internalImageBuildSpec}
}

func (imageBuildSpec *ImageBuildSpec) GetContainerImageFilePath() string {
	return imageBuildSpec.privateImageBuildSpec.ContainerImageFilePath
}

func (imageBuildSpec *ImageBuildSpec) GetBuildContextDir() string {
	return imageBuildSpec.privateImageBuildSpec.ContextDirPath
}

func (imageBuildSpec *ImageBuildSpec) GetTargetStage() string {
	return imageBuildSpec.privateImageBuildSpec.TargetStage
}

func (imageBuildSpec *ImageBuildSpec) GetBuildArgs() map[string]*string {
	return imageBuildSpec.privateImageBuildSpec.BuildArgs
}

func (imageBuildSpec *ImageBuildSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(imageBuildSpec.privateImageBuildSpec)
}

func (imageBuildSpec *ImageBuildSpec) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateImageBuildSpec{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	imageBuildSpec.privateImageBuildSpec = unmarshalledPrivateStructPtr
	return nil
}
