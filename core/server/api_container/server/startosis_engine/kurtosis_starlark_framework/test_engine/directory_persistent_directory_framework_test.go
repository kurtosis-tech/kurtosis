package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/stretchr/testify/require"
	"testing"
)

type directoryPersistentDirectoryTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestDirectoryPersistentDirectory() {
	suite.run(&directoryPersistentDirectoryTestCase{
		T: suite.T(),
	})
}

func (t *directoryPersistentDirectoryTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%d)", directory.DirectoryTypeName, directory.PersistentKeyAttr, testPersistentDirectoryKey, directory.SizeKeyAttr, testPersistentDirectorySize)
}

func (t *directoryPersistentDirectoryTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	directoryStarlark, ok := typeValue.(*directory.Directory)
	require.True(t, ok)

	artifactNames, found, err := directoryStarlark.GetArtifactNamesIfSet()
	require.Nil(t, err)
	require.False(t, found)
	require.Empty(t, artifactNames)

	persistentKey, found, err := directoryStarlark.GetPersistentKeyIfSet()
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, testPersistentDirectoryKey, persistentKey)

	size, err := directoryStarlark.GetSizeOrDefault()
	require.Nil(t, err)
	require.Equal(t, testPersistentDirectorySizeInBytes, size)
}
