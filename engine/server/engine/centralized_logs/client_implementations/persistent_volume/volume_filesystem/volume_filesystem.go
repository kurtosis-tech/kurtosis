package volume_filesystem

import (
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
)

// VolumeFilesystem interface is an abstraction of the disk filesystem
// primarily for the purpose of enabling unit testing persistentVolumeLogsDatabaseClient
type VolumeFilesystem interface {
	Open(name string) (VolumeFile, error)
	Create(name string) (VolumeFile, error)
	Stat(name string) (VolumeFileInfo, error)
	RemoveAll(path string) error
	Remove(filepath string) error
	Symlink(target, link string) error
	Walk(root string, walkFn filepath.WalkFunc) error
}

type VolumeFile interface {
	io.Reader
	Close() error
	WriteString(s string) (int, error)
}

type VolumeFileInfo interface {
	Mode() os.FileMode
}

// OsVolumeFilesystem is an implementation of the filesystem using disk
type OsVolumeFilesystem struct{}

func NewOsVolumeFilesystem() *OsVolumeFilesystem {
	return &OsVolumeFilesystem{}
}

func (fs *OsVolumeFilesystem) Open(name string) (VolumeFile, error) {
	return os.Open(name)
}

func (fs *OsVolumeFilesystem) Create(name string) (VolumeFile, error) {
	return os.Create(name)
}

func (fs *OsVolumeFilesystem) Stat(name string) (VolumeFileInfo, error) {
	return os.Stat(name)
}

func (fs *OsVolumeFilesystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fs *OsVolumeFilesystem) Remove(filepath string) error {
	return os.Remove(filepath)
}

func (fs *OsVolumeFilesystem) Symlink(target, link string) error {
	return os.Symlink(target, link)
}

func (fs *OsVolumeFilesystem) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// MockedVolumeFilesystem is an implementation used for unit testing
type MockedVolumeFilesystem struct {
	// uses an underlying map filesystem that's easy to mock file data with
	mapFS afero.Fs
}

func NewMockedVolumeFilesystem() *MockedVolumeFilesystem {
	return &MockedVolumeFilesystem{mapFS: afero.NewMemMapFs()}
}

func (fs *MockedVolumeFilesystem) Open(name string) (VolumeFile, error) {
	return fs.mapFS.Open(name)
}

func (fs *MockedVolumeFilesystem) Create(name string) (VolumeFile, error) {
	return fs.mapFS.Create(name)
}

func (fs *MockedVolumeFilesystem) Stat(name string) (VolumeFileInfo, error) {
	return fs.mapFS.Stat(name)
}

func (fs *MockedVolumeFilesystem) RemoveAll(path string) error {
	return fs.mapFS.RemoveAll(path)
}

func (fs *MockedVolumeFilesystem) Remove(filepath string) error {
	return fs.mapFS.Remove(filepath)
}

func (fs *MockedVolumeFilesystem) Symlink(target, link string) error {
	// afero.MemMapFs doesn't support symlinks so the best we can do is create the symlink
	_, err := fs.mapFS.Create(link)
	return err
}

func (fs *MockedVolumeFilesystem) Walk(root string, fn filepath.WalkFunc) error {
	return afero.Walk(fs.mapFS, root, fn)
}
