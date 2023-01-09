package uuid_generator

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
	"strings"
)

const (
	// Negative number indicates infinite replacements
	numberOfHyphenReplacements = -1

	// Length of shortened uuid
	shortenedUuidLength = 12
)

var (
	shortenedUuidRegex = regexp.MustCompile(fmt.Sprintf("[a-f0-9]{%v}", shortenedUuidLength))
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

func ShortenedUUIDString(fullUUID string) string {
	return fullUUID[:shortenedUuidLength]
}

func ISShortenedUUID(maybeShortenedUuid string) bool {
	return shortenedUuidRegex.MatchString(maybeShortenedUuid)
}
