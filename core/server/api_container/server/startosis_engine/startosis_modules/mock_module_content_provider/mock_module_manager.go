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

func NewEmptyMockModuleProvider() *MockModuleContentProvider {
	return NewMockModuleContentProvider(
		map[string]string{},
	)
}

func (moduleManager *MockModuleContentProvider) GetModuleContents(moduleID string) (string, error) {
	contents, found := moduleManager.modules[moduleID]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleID)
	}
	return contents, nil
}

func (moduleManager *MockModuleContentProvider) Add(moduleID string, contents string) {
	moduleManager.modules[moduleID] = contents
}
