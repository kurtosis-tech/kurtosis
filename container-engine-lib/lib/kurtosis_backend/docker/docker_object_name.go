package docker

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	dockerObjectNameRegexStr = "^[a-zA-Z0-9_-.]+$"
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
	return &DockerObjectName{value: str}, nil
}
func (key *DockerObjectName) GetString() string {
	return key.value
}

