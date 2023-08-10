package file_artifacts_db

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFileArtifactPersistance(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	fileArtifactDb, err := GetFileArtifactsDbForTesting(enclaveDb, map[string]string{})
	require.Nil(t, err)
	require.Empty(t, fileArtifactDb.data.ArtifactNameToArtifactUuid)
	require.Empty(t, fileArtifactDb.data.ArtifactContentMd5)
	require.Empty(t, fileArtifactDb.data.ShortenedUuidToFullUuid)
	fileArtifactDb.SetArtifactUuid("1", "1")
	fileArtifactDb.SetContentMd5("1", []byte("1"))
	fileArtifactDb.SetFullUuid("1", []string{"1"})
	require.Nil(t, fileArtifactDb.Persist())
	fileArtifactDb, err = getFileArtifactsDbFromEnclaveDb(enclaveDb, &fileArtifactData{
		map[string]string{},
		map[string][]string{},
		map[string][]byte{},
	})
	require.Nil(t, err)
	require.Len(t, fileArtifactDb.GetArtifactUuidMap(), 1)
	require.Len(t, fileArtifactDb.GetFullUuidMap(), 1)
	require.Len(t, fileArtifactDb.GetContentMd5Map(), 1)
}
