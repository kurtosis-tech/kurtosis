package module_manager

import "github.com/kurtosis-tech/stacktrace"

type MockPackageManager struct {
	packages map[string]string
}

func NewPackageManager(seedPackages map[string]string) *MockPackageManager {
	return &MockPackageManager{
		packages: seedPackages,
	}
}

func (mockPackageManager *MockPackageManager) GetModule(packageURL string) (string, error) {
	contents, found := mockPackageManager.packages[packageURL]
	if !found {
		return "", stacktrace.NewError("Package '%v' not found", packageURL)
	}
	return contents, nil
}
