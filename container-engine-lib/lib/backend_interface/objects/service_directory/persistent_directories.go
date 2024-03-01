package service_directory

import "regexp"

const (
	// PersistentKeyRegex implements RFC-1035 for naming persistent directory keys, namely:
	// * contain at most 63 characters
	// * contain only lowercase alphanumeric characters or '-'
	// * start with an alphabetic character
	// * end with an alphanumeric character
	// This is in order to stick to the 1035 standard which we enforce for all objects created
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	PersistentKeyRegex            = "[a-z]([-a-z0-9]{0,61}[a-z0-9])?"
	WordWrappedPersistentKeyRegex = "^" + PersistentKeyRegex + "$"
)

var (
	compiledWordWrappedPersistentKeyRegex = regexp.MustCompile(WordWrappedPersistentKeyRegex)
)

type DirectoryPersistentKey string
type DirectoryPersistentSize int64

type PersistentDirectory struct {
	PersistentKey DirectoryPersistentKey
	Size          DirectoryPersistentSize
}

type PersistentDirectories struct {
	ServiceDirpathToPersistentDirectory map[string]PersistentDirectory
}

func NewPersistentDirectories(persistentDirectories map[string]PersistentDirectory) *PersistentDirectories {
	return &PersistentDirectories{
		ServiceDirpathToPersistentDirectory: persistentDirectories,
	}
}

func IsPersistentKeyValid(persistentKey DirectoryPersistentKey) bool {
	return compiledWordWrappedPersistentKeyRegex.MatchString(string(persistentKey))
}
