package startosis_engine

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveIfPresent(t *testing.T) {
	imageCurrentlyBeingDownloaded := []string{
		"kurtosistech/image_1",
		"kurtosistech/image_2",
	}
	result := removeIfPresent(imageCurrentlyBeingDownloaded, "kurtosistech/image_2")
	expectedImageCurrentlyBeingDownloaded := []string{
		"kurtosistech/image_1",
	}
	require.Equal(t, expectedImageCurrentlyBeingDownloaded, result)
}

func TestRemoveIfPresent_NotPresent(t *testing.T) {
	imageCurrentlyBeingDownloaded := []string{
		"kurtosistech/image_1",
		"kurtosistech/image_2",
	}
	result := removeIfPresent(imageCurrentlyBeingDownloaded, "kurtosistech/image_3")
	expectedImageCurrentlyBeingDownloaded := []string{
		"kurtosistech/image_1",
		"kurtosistech/image_2",
	}
	require.Equal(t, expectedImageCurrentlyBeingDownloaded, result)
}
