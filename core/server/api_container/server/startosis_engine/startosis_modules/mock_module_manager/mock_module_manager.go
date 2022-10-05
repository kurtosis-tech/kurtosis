package mock_module_manager

import "github.com/kurtosis-tech/stacktrace"

type MockModuleManager struct {
	modules map[string]string
}

func NewMockModuleManager(seedModules map[string]string) *MockModuleManager {
	return &MockModuleManager{
		modules: seedModules,
	}
}

func (moduleManager *MockModuleManager) GetModule(moduleID string) (string, error) {
	contents, found := moduleManager.modules[moduleID]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleID)
	}
	return contents, nil
}
