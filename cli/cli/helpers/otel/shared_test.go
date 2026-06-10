package otel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLokiSinkDisablesHealthcheck(t *testing.T) {
	lokiHost := "http://collector:3500"

	sinks := NewLokiSink(lokiHost)

	lokiSink, found := sinks["loki"]
	require.True(t, found)
	require.Equal(t, "loki", lokiSink["type"])
	require.Equal(t, lokiHost, lokiSink["endpoint"])
	require.Equal(t, map[string]bool{"enabled": false}, lokiSink["healthcheck"])
}

func TestWriteTempFileUsesWorldReadablePermissions(t *testing.T) {
	filepath, err := writeTempFile("otel-test-*.txt", "test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Remove(filepath))
	})

	fileInfo, err := os.Stat(filepath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(tempFilePerms), fileInfo.Mode().Perm())
}
