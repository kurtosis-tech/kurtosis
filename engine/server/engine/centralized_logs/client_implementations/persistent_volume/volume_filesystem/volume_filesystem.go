package volume_filesystem

import (
	"io"
	"os"
	"strings"
	"testing/fstest"
)

const (
	forwardSlash = "/"
)

// VolumeFilesystem interface is an abstraction of the disk filesystem
// primarily for the purpose of enabling unit testing persistentVolumeLogsDatabaseClient
type VolumeFilesystem interface {
	Open(name string) (VolumeFile, error)
	Stat(name string) (VolumeFileInfo, error)
	RemoveAll(path string) error
}

type VolumeFile interface {
	io.Reader
	Close() error
}

type VolumeFileInfo interface {
}

// OsVolumeFilesystem is an implementation of the filesystem using disk
type OsVolumeFilesystem struct{}

func NewOsVolumeFilesystem() *OsVolumeFilesystem {
	return &OsVolumeFilesystem{}
}

func (fs *OsVolumeFilesystem) Open(name string) (VolumeFile, error) {
	return os.Open(name)
}

func (fs *OsVolumeFilesystem) Stat(name string) (VolumeFileInfo, error) {
	return os.Stat(name)
}

func (fs *OsVolumeFilesystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// MockedVolumeFilesystem is an implementation used for unit testing
type MockedVolumeFilesystem struct {
	// we use an underlying map filesystem that's easy to mock file data with
	mapFS *fstest.MapFS
}

func NewMockedVolumeFilesystem(fs *fstest.MapFS) *MockedVolumeFilesystem {
	return &MockedVolumeFilesystem{mapFS: fs}
}

func (fs *MockedVolumeFilesystem) Open(name string) (VolumeFile, error) {
	// Trim any forward slashes from this filepath
	// fstest.MapFS doesn't like absolute paths!!!
	return fs.mapFS.Open(trimForwardSlash(name))
}

func (fs *MockedVolumeFilesystem) Stat(name string) (VolumeFileInfo, error) {
	// Trim any forward slashes from this filepath
	// fstest.MapFS doesn't like absolute paths!!!
	return fs.mapFS.Stat(trimForwardSlash(name))
}

func (fs *MockedVolumeFilesystem) RemoveAll(path string) error {
	path = trimForwardSlash(path)
	for filepath := range *fs.mapFS {
		if strings.HasPrefix(filepath, path) {
			delete(*fs.mapFS, filepath)
		}
	}
	return nil
}

func trimForwardSlash(name string) string {
	return strings.TrimLeft(name, forwardSlash)
}
