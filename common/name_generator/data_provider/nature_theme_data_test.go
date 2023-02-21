package data_provider

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_convertMapSetToStringArray(t *testing.T) {
	data := map[string]bool{
		"test":             true,
		"test_key_another": true,
	}

	actual := convertMapSetToStringArray(data)

	require.Contains(t, actual, "test")
	require.Contains(t, actual, "test_key_another")
	require.NotContains(t, actual, "abc")
	require.Len(t, actual, len(data))
}
