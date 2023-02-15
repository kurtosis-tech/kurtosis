package uuid_generator

import (
	"github.com/google/uuid"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	// Negative number indicates infinite replacements
	numberOfHyphenReplacements = -1

	// Length of shortened uuid
	shortenedUuidLength = 12
)

// Generates a UUID with the dashes removed
func GenerateUUIDString() (string, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating a new UUID")
	}

	result := strings.Replace(newUUID.String(), "-", "", numberOfHyphenReplacements)
	return result, nil
}

// IsUUID checks whether a given string is a UUID
func IsUUID(maybeUuid string) bool {
	_, err := uuid.Parse(maybeUuid)
	return err == nil
}

// ShortenedUUIDString returns at most the first 12 characters of the uuid
// Though a valid uuid is 12 characters, we need to support older shorter uuids for backwards compatibility
func ShortenedUUIDString(fullUUID string) string {
	lengthToTrim := shortenedUuidLength
	if lengthToTrim > len(fullUUID) {
		logrus.Warnf("Encountered a uuid '%v' that is shorter than expected", fullUUID)
		lengthToTrim = len(fullUUID)
	}
	return fullUUID[:lengthToTrim]
}
