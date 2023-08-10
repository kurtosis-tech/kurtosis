package service_directory

type DirectoryPersistentKey string

type PersistentDirectories struct {
	ServiceDirpathToDirectoryPersistentKey map[string]DirectoryPersistentKey
}

func NewPersistentDirectories(serviceDirpathToDirectoryPersistentKey map[string]DirectoryPersistentKey) *PersistentDirectories {
	return &PersistentDirectories{
		ServiceDirpathToDirectoryPersistentKey: serviceDirpathToDirectoryPersistentKey,
	}
}
