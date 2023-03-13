package docker_object_name

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	dockerObjectNameRegexStr = "^[a-zA-Z0-9-._]+$"

	// We couldn't find any actual limit, but this is very sensible
	maxLength = 256
)
var dockerObjectNameRegex = regexp.MustCompile(dockerObjectNameRegexStr)

// Represents a Docker label that is guaranteed to be valid for the Docker engine
// NOTE: This is a struct-based enum
type DockerObjectName struct {
	value string
}
func CreateNewDockerObjectName(str string) (*DockerObjectName, error) {
	if !dockerObjectNameRegex.MatchString(str) {
		return nil, stacktrace.NewError("Object name '%v' doesn't match Docker docker object name regex '%v'", str, dockerObjectNameRegexStr)
	}


	if len(str) > maxLength {
		return nil, stacktrace.NewError("Object name string '%v' is longer than max allowed object name length '%v'", str, maxLength)
	}

	return &DockerObjectName{value: str}, nil
}
func (key *DockerObjectName) GetString() string {
	return key.value
}

