package kubernetes_files_artifact_expansion

type ExpandSingleArtifactConfig struct {

}

type ExpandFilesArtifactsConfig struct {
	// The image that will be run as an InitContainer before the user service starts
	expanderImage string

	// The environment variables that the expander container will be passed in, to configure
	// its operation
	expanderEnvVars map[string]string


}
