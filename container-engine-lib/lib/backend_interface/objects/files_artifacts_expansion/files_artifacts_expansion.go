package files_artifacts_expansion

type FilesArtifactsExpansion struct {
	// The image that will run before the user service starts, to expand files artifacts into volumes
	// so that the user service container has what it expects
	ExpanderImage string

	// The environment variables that the expander container will be passed in, to configure
	// its operation
	ExpanderEnvVars map[string]string

	// Map of dirpaths that the expander container expects (which the expander will expand into), mapped to
	// dirpaths on the user service container where those same directories should be made available
	ExpanderDirpathsToServiceDirpaths map[string]string
}