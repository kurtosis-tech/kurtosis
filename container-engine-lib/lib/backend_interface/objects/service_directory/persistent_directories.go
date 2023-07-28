package service_directory

type DirectoryPersistentKey string

type PersistentDirectories struct {
	ServiceDirpathToDirectoryPersistentKey map[string]DirectoryPersistentKey
}
