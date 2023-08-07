package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/stretchr/testify/require"
	"testing"
)

type directoryFileArtifactTestCase struct {
	*testing.T
}

func newDirectoryFileArtifactTestCase(t *testing.T) *directoryFileArtifactTestCase {
	return &directoryFileArtifactTestCase{
		T: t,
	}
}

func (t *directoryFileArtifactTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", directory.DirectoryTypeName, "FileArtifact")
}

func (t *directoryFileArtifactTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.ArtifactNameAttr, TestFilesArtifactName1)
}

func (t *directoryFileArtifactTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	directoryStarlark, ok := typeValue.(*directory.Directory)
	require.True(t, ok)

	artifactName, found, err := directoryStarlark.GetArtifactNameIfSet()
	require.Nil(t, err)
	require.True(t, found)
	require.Equal(t, TestFilesArtifactName1, artifactName)

	persistentKey, found, err := directoryStarlark.GetPersistentKeyIfSet()
	require.Nil(t, err)
	require.False(t, found)
	require.Empty(t, persistentKey)
}
