package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/stretchr/testify/require"
	"testing"
)

type directoryMultipleFileArtifactsTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestDirectoryMultipleFileArtifacts() {
	suite.run(&directoryMultipleFileArtifactsTestCase{
		T: suite.T(),
	})
}

func (t *directoryMultipleFileArtifactsTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=[%q, %q])", directory.DirectoryTypeName, directory.ArtifactNamesAttr, testFilesArtifactName1, testFilesArtifactName2)
}

func (t *directoryMultipleFileArtifactsTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	directoryStarlark, ok := typeValue.(*directory.Directory)
	require.True(t, ok)

	artifactNames, found, err := directoryStarlark.GetArtifactNamesIfSet()
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, []string{testFilesArtifactName1, testFilesArtifactName2}, artifactNames)
}
