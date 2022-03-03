package docker_label_key

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	dockerLabelKeyRegexStr = "^[a-z0-9-.]+$"
)
var dockerLabelKeyRegex = regexp.MustCompile(dockerLabelKeyRegexStr)

// TODO MAKE THIS AN INTERFACE????

// Represents a Docker label that is guaranteed to be valid for the Docker engine
// NOTE: This is a struct-based enum
type DockerLabelKey struct {
	value string
}
// NOTE: This is ONLY for areas where the label is declared statically!! Any sort of dynamic/runtime label creation
//  should use CreateNewDockerLabelKey
func MustCreateNewDockerLabelKey(str string) *DockerLabelKey {
	key, err := CreateNewDockerLabelKey(str)
	if err != nil {
		panic(err)
	}
	return key
}
func CreateNewDockerLabelKey(str string) (*DockerLabelKey, error) {
	if !dockerLabelKeyRegex.MatchString(str) {
		return nil, stacktrace.NewError("Label key string '%v' doesn't match Docker label key regex '%v'", str, dockerLabelKeyRegexStr)
	}
	return &DockerLabelKey{value: str}, nil
}
func (key *DockerLabelKey) GetString() string {
	return key.value
}

