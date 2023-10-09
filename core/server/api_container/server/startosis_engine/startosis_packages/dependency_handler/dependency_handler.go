package dependency_handler

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/package_replace_options_repository"
)

type DependencyHandler struct {
	packageReplaceOptionsRepository *package_replace_options_repository.PackageReplaceOptionsRepository
	gitPackageContentProvider       *git_package_content_provider.GitPackageContentProvider
}

func NewDependencyHandler(packageReplaceOptionsRepository *package_replace_options_repository.PackageReplaceOptionsRepository, gitPackageContentProvider *git_package_content_provider.GitPackageContentProvider) *DependencyHandler {
	return &DependencyHandler{packageReplaceOptionsRepository: packageReplaceOptionsRepository, gitPackageContentProvider: gitPackageContentProvider}
}
