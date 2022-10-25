package mock_module_content_provider

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	unimplementedMessage = "This method isn't implemented!!!!"
)

type MockModuleContentProvider struct {
	modules map[string]string
}

func NewMockModuleContentProvider(seedModules map[string]string) *MockModuleContentProvider {
	mapToTempFile := make(map[string]string, len(seedModules))
	for key, value := range seedModules {
		mapToTempFile[key] = writeContentToTempFileOrPanic(value)
	}
	return &MockModuleContentProvider{
		modules: mapToTempFile,
	}
}

func NewEmptyMockModuleContentProvider() *MockModuleContentProvider {
	return NewMockModuleContentProvider(
		map[string]string{},
	)
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
	if fileContent, err := os.ReadFile(absFilePath); err == nil {
		return string(fileContent), nil
	}
	return "", stacktrace.NewError("Unable to read content of module '%v'", moduleID)
}

func (provider *MockModuleContentProvider) AddFileContent(moduleID string, contents string) {
	provider.modules[moduleID] = writeContentToTempFileOrPanic(contents)
}

func writeContentToTempFileOrPanic(fileContent string) string {
	tempFile, err := os.CreateTemp("", "mock_module_content_provider_*")
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to create temp file for MockModuleContentProvider"))
	}
	_, err = tempFile.WriteString(fileContent)
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to write content to temp file for MockModuleContentProvider"))
	}
	return tempFile.Name()
}
