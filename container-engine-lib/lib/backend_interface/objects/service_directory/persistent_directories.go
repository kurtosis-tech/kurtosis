package service_directory

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
