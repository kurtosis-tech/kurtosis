package services

import (
	"path/filepath"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
type SharedPath struct {
	//Absolute path in the container where this code is running
	absPathOnThisContainer string
	//Absolute path in the service container
	absPathOnServiceContainer string
}

func NewSharedPath(absPathOnThisContainer string, absPathOnServiceContainer string) *SharedPath {
	return &SharedPath{absPathOnThisContainer: absPathOnThisContainer, absPathOnServiceContainer: absPathOnServiceContainer}
}
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
func (s SharedPath) GetAbsPathOnThisContainer() string {
	return s.absPathOnThisContainer
}
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
func (s SharedPath) GetAbsPathOnServiceContainer() string {
	return s.absPathOnServiceContainer
}

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
func (s SharedPath) GetChildPath(pathElement string) *SharedPath {

	absPathOnThisContainer := filepath.Join(s.absPathOnThisContainer, pathElement)

	absPathOnServiceContainer := filepath.Join(s.absPathOnServiceContainer, pathElement)

	sharedPath := NewSharedPath(absPathOnThisContainer, absPathOnServiceContainer)

	return sharedPath
}
