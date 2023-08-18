package persistent_volume

import (
	"io"
	"os"
	"testing/fstest"
)

// VolumeFilesystem interface is an abstraction of the disk filesystem
// primarily for the purpose of enabling unit testing PersistentVolumeLogsDatabaseClient
type VolumeFilesystem interface {
	Open(name string) (VolumeFile, error)
}

type VolumeFile interface {
	io.Reader
}

// OsVolumeFilesystem is an implementation of the filesystem using disk
type OsVolumeFilesystem struct{}

func NewOsVolumeFilesystem() *OsVolumeFilesystem {
	return &OsVolumeFilesystem{}
}

func (fs *OsVolumeFilesystem) Open(name string) (VolumeFile, error) { return os.Open(name) }

// MockedVolumeFilesystem is an implementation used for unit testing
type MockedVolumeFilesystem struct {
	// we use an underlying map filesystem that's easy to mock file data with
	mapFS fstest.MapFS
}

func NewMockedVolumeFilesystem(fs fstest.MapFS) *MockedVolumeFilesystem {
	return &MockedVolumeFilesystem{mapFS: fs}
}

func (fs *MockedVolumeFilesystem) Open(name string) (VolumeFile, error) { return fs.mapFS.Open(name) }
