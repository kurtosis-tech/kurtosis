package mock_module_content_provider

import (
	"github.com/kurtosis-tech/stacktrace"
)

type MockModuleContentProvider struct {
	modules map[string]string
}

func NewMockModuleContentProvider(seedModules map[string]string) *MockModuleContentProvider {
	return &MockModuleContentProvider{
		modules: seedModules,
	}
}

func NewEmptyMockModuleContentProvider() *MockModuleContentProvider {
	return NewMockModuleContentProvider(
		map[string]string{},
	)
}

func (provider *MockModuleContentProvider) GetModuleContents(moduleID string) (string, error) {
	contents, found := provider.modules[moduleID]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleID)
	}
	return contents, nil
}

func (provider *MockModuleContentProvider) Add(moduleID string, contents string) {
	provider.modules[moduleID] = contents
}
