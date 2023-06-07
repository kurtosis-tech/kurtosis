package fluentd

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderConfig(t *testing.T) {
	config := newDefaultFluentdConfigForKurtosisCentralizedLogs(9880)

	renderedConfig, err := config.RenderConfig()
	require.NoError(t, err)
	require.NotNil(t, renderedConfig)
}
