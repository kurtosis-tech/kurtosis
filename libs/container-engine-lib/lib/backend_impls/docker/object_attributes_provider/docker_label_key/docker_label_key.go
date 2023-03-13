package docker_label_key

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	dockerLabelKeyRegexStr = "^[a-z0-9-.]+$"

	// It doesn't seem Docker actually has a label key length limit, but we implement one of our own for practicality
	maxLabelLength = 256
)
var dockerLabelKeyRegex = regexp.MustCompile(dockerLabelKeyRegexStr)

// Represents a Docker label that is guaranteed to be valid for the Docker engine
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
	if len(str) > maxLabelLength {
		return nil, stacktrace.NewError("Label key string '%v' is longer than max label key length '%v'", str, maxLabelLength)
	}
	return &DockerLabelKey{value: str}, nil
}
func (key *DockerLabelKey) GetString() string {
	return key.value
}

