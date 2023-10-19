package docker_label_key

import (
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	dockerLabelKeyRegexStr = "^[a-z0-9-._]+$"

	// It doesn't seem Docker actually has a label key length limit, but we implement one of our own for practicality
	maxLabelLength = 256
)

var dockerLabelKeyRegex = regexp.MustCompile(dockerLabelKeyRegexStr)

// Represents a Docker label that is guaranteed to be valid for the Docker engine
type DockerLabelKey struct {
	value string
}

// NOTE: This is ONLY for areas where the label is declared statically!! Any sort of dynamic/runtime label creation
// should use createNewDockerLabelKey
func MustCreateNewDockerLabelKey(str string) *DockerLabelKey {
	key, err := createNewDockerLabelKey(str)
	if err != nil {
		panic(err)
	}
	return key
}

// CreateNewDockerUserCustomLabelKey creates a custom uer Docker label with the Kurtosis custom user prefix
func CreateNewDockerUserCustomLabelKey(str string) (*DockerLabelKey, error) {
	if str == "" || str == " " {
		return nil, stacktrace.NewError("Received an empty user custom label key")
	}
	labelKeyStr := customUserLabelsKeyPrefixStr + str
	return createNewDockerLabelKey(labelKeyStr)
}

func createNewDockerLabelKey(str string) (*DockerLabelKey, error) {
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
