package uuid_generator

import (
	"github.com/google/uuid"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	// Negative number indicates infinite replacements
	numberOfHyphenReplacements = -1
)

// Generates a UUID with the dashes removed
func GenerateUUIDString() (string, error) {
	engineUuid, err := uuid.NewUUID()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating a UUID for the engine GUID")
	}

	result := strings.Replace(engineUuid.String(), "-", "", numberOfHyphenReplacements)
	return result, nil
}
