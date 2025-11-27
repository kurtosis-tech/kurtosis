package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

type serviceConfigMultipleFilesInSameFolderTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigMultipleFilesInSameFolder() {

	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	suite.run(&serviceConfigMultipleFilesInSameFolderTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigMultipleFilesInSameFolderTestCase) GetStarlarkCode() string {
	filesArtifactsDirectory := fmt.Sprintf("%s(%s=[%q, %q])", directory.DirectoryTypeName, directory.ArtifactNamesAttr, testFilesArtifactName1, testFilesArtifactName2)
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.FilesAttr, fmt.Sprintf("{%q: %s}", testFilesArtifactPath1, filesArtifactsDirectory),
	)

	return starlarkCode
}

func (t *serviceConfigMultipleFilesInSameFolderTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModulePackageId,
		testModuleMainFileLocator,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, err)

	require.Equal(t, testContainerImageName, serviceConfig.GetContainerImageName())

	expectedFilesArtifactMap := map[string][]string{
		testFilesArtifactPath1: {testFilesArtifactName1, testFilesArtifactName2},
	}
	require.NotNil(t, serviceConfig.GetFilesArtifactsExpansion())
	require.Equal(t, expectedFilesArtifactMap, serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers)
}
