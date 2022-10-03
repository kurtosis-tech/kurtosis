package module_manager

import "github.com/kurtosis-tech/stacktrace"

type MockModuleManager struct {
	packages map[string]string
}

func NewPackageManager(seedPackages map[string]string) *MockModuleManager {
	return &MockModuleManager{
		packages: seedPackages,
	}
}

func (mockPackageManager *MockModuleManager) GetModule(moduleURL string) (string, error) {
	contents, found := mockPackageManager.packages[moduleURL]
	if !found {
		return "", stacktrace.NewError("Module '%v' not found", moduleURL)
	}
	return contents, nil
}
