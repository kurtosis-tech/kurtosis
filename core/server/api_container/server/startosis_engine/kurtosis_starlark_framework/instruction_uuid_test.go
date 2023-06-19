package kurtosis_starlark_framework

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateInstructionUuid(t *testing.T) {
	uuid, err := GenerateInstructionUuid()
	require.NoError(t, err)
	require.Regexp(t, "[a-f0-9]{32}", uuid)
}
