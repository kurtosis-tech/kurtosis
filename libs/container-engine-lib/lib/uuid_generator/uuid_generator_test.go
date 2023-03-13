package uuid_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateUUIDString(t *testing.T) {
	uuid, err := GenerateUUIDString()
	require.Nil(t, err)
	require.Len(t, uuid, 32)
	require.True(t, IsUUID(uuid))
	shortenedUuid := ShortenedUUIDString(uuid)
	require.Len(t, shortenedUuid, shortenedUuidLength)
}

func TestUUIDBackwardsCompatibility(t *testing.T) {
	oldUuid := "short-uuid"
	require.Equal(t, oldUuid, ShortenedUUIDString(oldUuid))
}
