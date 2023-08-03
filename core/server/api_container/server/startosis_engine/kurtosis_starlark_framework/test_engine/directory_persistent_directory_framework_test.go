package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/stretchr/testify/require"
	"testing"
)

type directoryPersistentDirectoryTestCase struct {
	*testing.T
}

func newDirectoryPersistnetDirectoryTestCase(t *testing.T) *directoryPersistentDirectoryTestCase {
	return &directoryPersistentDirectoryTestCase{
		T: t,
	}
}

func (t *directoryPersistentDirectoryTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", directory.DirectoryTypeName, "PersistentDirectory")
}

func (t *directoryPersistentDirectoryTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.PersistentKeyAttr, TestPersistentDirectoryKey)
}

func (t *directoryPersistentDirectoryTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	directoryStarlark, ok := typeValue.(*directory.Directory)
	require.True(t, ok)

	artifactName, found, err := directoryStarlark.GetArtifactNameIfSet()
	require.Nil(t, err)
	require.False(t, found)
	require.Empty(t, artifactName)

	persistentKey, found, err := directoryStarlark.GetPersistentKeyIfSet()
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, TestPersistentDirectoryKey, persistentKey)
}
