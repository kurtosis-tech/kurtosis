package mock_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	unimplementedMessage = "This method isn't implemented!!!!"

	defaultTempDir          = ""
	tempProviderFilePattern = "mock_module_content_provider_*"
)

type MockPackageContentProvider struct {
	starlarkPackages map[string]string
}

func NewMockModuleContentProvider() *MockPackageContentProvider {
	return &MockPackageContentProvider{
		starlarkPackages: make(map[string]string),
	}
}

func (provider *MockPackageContentProvider) GetOnDiskAbsoluteFilePath(packageId string) (string, *startosis_errors.InterpretationError) {
	absFilePath, found := provider.starlarkPackages[packageId]
	if !found {
		return "", startosis_errors.NewInterpretationError("Module '%v' not found", packageId)
	}
	if _, err := os.Stat(absFilePath); err != nil {
		return "", startosis_errors.NewInterpretationError("Unable to read content of module '%v'", packageId)
	}
	return absFilePath, nil
}

func (provider *MockPackageContentProvider) ClonePackage(_ string) (string, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)
}

func (provider *MockPackageContentProvider) StorePackageContents(string, []byte, bool) (string, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)
}

func (provider *MockPackageContentProvider) GetPackageContents(packageId string) (string, *startosis_errors.InterpretationError) {
	absFilePath, found := provider.starlarkPackages[packageId]
	if !found {
		return "", startosis_errors.NewInterpretationError("Package '%v' not found", packageId)
	}
	fileContent, err := os.ReadFile(absFilePath)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("Unable to read content of package '%v'", packageId)
	}
	return string(fileContent), nil
}

func (provider *MockPackageContentProvider) AddFileContent(packageId string, contents string) error {
	absFilePath, err := writeContentToTempFile(contents)
	if err != nil {
		return stacktrace.Propagate(err, "Error writing content to temporary file")
	}
	provider.starlarkPackages[packageId] = absFilePath
	return nil
}

func (provider *MockPackageContentProvider) BulkAddFileContent(packageIdToContent map[string]string) error {
	for moduleId, moduleContent := range packageIdToContent {
		absFilePath, err := writeContentToTempFile(moduleContent)
		if err != nil {
			return stacktrace.Propagate(err, "Error writing content of module '%s' to temporary file", moduleId)
		}
		provider.starlarkPackages[moduleId] = absFilePath
	}
	return nil
}

func (provider *MockPackageContentProvider) RemoveAll() map[string]error {
	deletionErrors := make(map[string]error)
	for moduleId, absFilePath := range provider.starlarkPackages {
		if err := os.Remove(absFilePath); err != nil {
			deletionErrors[moduleId] = err
		}
	}
	if len(deletionErrors) > 0 {
		return deletionErrors
	}
	return nil
}

func writeContentToTempFile(fileContent string) (string, error) {
	tempFile, err := os.CreateTemp(defaultTempDir, tempProviderFilePattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to create temp file for MockModuleContentProvider")
	}
	_, err = tempFile.WriteString(fileContent)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to write content to temp file '%v' for MockModuleContentProvider", tempFile.Name())
	}
	return tempFile.Name(), nil
}
