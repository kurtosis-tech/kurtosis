/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_expander_consts

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	DockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the suite execution volume (which contains the artifacts)
	//  will be mounted
	SuiteExecutionVolumeMountDirpath = "/suite-execution"

	// Dirpath on the artifact expander container where the destination volume will be mounted
	DestinationVolumeMountDirpath = "/dest"
)

// Image-specific generator of the command that should be run to extract the artifact at the given filepath
//  to the destination
func GetExtractionCommand(artifactFilepath string) (dockerRunCmd []string) {
	return []string{
		"tar",
		"-xzvf",
		artifactFilepath,
		"-C",
		DestinationVolumeMountDirpath,
	}
}
