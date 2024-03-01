package mock_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/yaml_parser"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"os"
	"strings"
)

const (
	unimplementedMessage = "This method isn't implemented!!!!"

	defaultTempDir          = ""
	tempProviderFilePattern = "mock_module_content_provider_*"
)

// MockPackageContentProvider is mock for PackageContentProvider interface
//
// TODO: use the mockery-generated mock: startosis_package.MockPackageContentProvider
type MockPackageContentProvider struct {
	starlarkPackages map[string]string
	packageId        string
}

func NewMockPackageContentProvider() *MockPackageContentProvider {
	return &MockPackageContentProvider{
		starlarkPackages: make(map[string]string),
		packageId:        "",
	}
}

func (provider *MockPackageContentProvider) GetOnDiskAbsolutePackageFilePath(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError) {
	absFilePath, found := provider.starlarkPackages[fileInsidePackageUrl]
	if !found {
		return "", startosis_errors.NewInterpretationError("Module '%v' not found", fileInsidePackageUrl)
	}
	if _, err := os.Stat(absFilePath); err != nil {
		return "", startosis_errors.NewInterpretationError("Unable to read content of package '%v'", fileInsidePackageUrl)
	}
	return absFilePath, nil
}

func (provider *MockPackageContentProvider) GetOnDiskAbsolutePath(repositoryPathURL string) (string, *startosis_errors.InterpretationError) {
	absFilePath, found := provider.starlarkPackages[repositoryPathURL]
	if !found {
		return "", startosis_errors.NewInterpretationError("Module '%v' not found", repositoryPathURL)
	}
	if _, err := os.Stat(absFilePath); err != nil {
		return "", startosis_errors.NewInterpretationError("Unable to read content of package '%v'", repositoryPathURL)
	}
	return absFilePath, nil
}

func (provider *MockPackageContentProvider) ClonePackage(_ string) (string, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)
}

func (provider *MockPackageContentProvider) GetOnDiskAbsolutePackagePath(packageId string) (string, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)

}

func (provider *MockPackageContentProvider) StorePackageContents(_ string, _ io.Reader, _ bool) (string, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)
}

func (provider *MockPackageContentProvider) GetKurtosisYaml(packageAbsolutePathOnDisk string) (*yaml_parser.KurtosisYaml, *startosis_errors.InterpretationError) {
	panic(unimplementedMessage)
}

func (provider *MockPackageContentProvider) CloneReplacedPackagesIfNeeded(currentPackageReplaceOptions map[string]string) *startosis_errors.InterpretationError {
	return nil
}

func (provider *MockPackageContentProvider) GetModuleContents(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError) {
	absFilePath, found := provider.starlarkPackages[fileInsidePackageUrl]
	if !found {
		return "", startosis_errors.NewInterpretationError("Package '%v' not found", fileInsidePackageUrl)
	}
	fileContent, err := os.ReadFile(absFilePath)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("Unable to read content of package '%v'", fileInsidePackageUrl)
	}
	return string(fileContent), nil
}

func (provider *MockPackageContentProvider) GetAbsoluteLocator(packageId string, parentModuleId string, relativeOrAbsoluteModulePath string, packageReplaceOptions map[string]string) (string, *startosis_errors.InterpretationError) {
	if strings.HasPrefix(relativeOrAbsoluteModulePath, parentModuleId) {
		return "", startosis_errors.NewInterpretationError("Cannot use local absolute locators")
	}

	if strings.HasPrefix(relativeOrAbsoluteModulePath, shared_utils.GithubDomainPrefix) {
		return relativeOrAbsoluteModulePath, nil
	}
	return provider.packageId, nil
}

func (provider *MockPackageContentProvider) AddFileContent(packageId string, contents string) error {
	absFilePath, err := writeContentToTempFile(contents)
	if err != nil {
		return stacktrace.Propagate(err, "Error writing content to temporary file")
	}
	provider.starlarkPackages[packageId] = absFilePath
	provider.packageId = packageId
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
