package mock_module_content_provider

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	unimplementedMessage = "This method isn't implemented!!!!"

	defaultTempDir          = ""
	tempProviderFilePattern = "mock_module_content_provider_*"
)

type MockModuleContentProvider struct {
	modules map[string]string
}

func NewMockModuleContentProvider() *MockModuleContentProvider {
	return &MockModuleContentProvider{
		modules: make(map[string]string),
	}
}

func (provider *MockModuleContentProvider) GetOnDiskAbsoluteFilePath(moduleID string) (string, error) {
	absFilePath, found := provider.modules[moduleID]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleID)
	}
	if _, err := os.Stat(absFilePath); err != nil {
		return "", stacktrace.NewError("Unable to read content of module '%v'", moduleID)
	}
	return absFilePath, nil
}

func (provider *MockModuleContentProvider) StoreModuleContents(string, []byte, bool) (string, error) {
	panic(unimplementedMessage)
}

func (provider *MockModuleContentProvider) GetModuleContents(moduleID string) (string, error) {
	absFilePath, found := provider.modules[moduleID]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleID)
	}
	fileContent, err := os.ReadFile(absFilePath)
	if err != nil {
		return "", stacktrace.NewError("Unable to read content of module '%v'", moduleID)
	}
	return string(fileContent), nil
}

func (provider *MockModuleContentProvider) AddFileContent(moduleID string, contents string) error {
	absFilePath, err := writeContentToTempFile(contents)
	if err != nil {
		return stacktrace.Propagate(err, "Error writing content to temporary file")
	}
	provider.modules[moduleID] = absFilePath
	return nil
}

func (provider *MockModuleContentProvider) BulkAddFileContent(moduleIdToContent map[string]string) error {
	for moduleId, moduleContent := range moduleIdToContent {
		absFilePath, err := writeContentToTempFile(moduleContent)
		if err != nil {
			return stacktrace.Propagate(err, "Error writing content of module '%s' to temporary file", moduleId)
		}
		provider.modules[moduleId] = absFilePath
	}
	return nil
}

func (provider *MockModuleContentProvider) RemoveAll() map[string]error {
	deletionErrors := make(map[string]error)
	for moduleId, absFilePath := range provider.modules {
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
