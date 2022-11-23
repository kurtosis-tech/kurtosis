package exec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateArgs_valid(t *testing.T) {
	input := `{"hello": "world!"}`
	err := validateModuleArgs(input)
	require.Nil(t, err)
}

func TestValidateArgs_invalid(t *testing.T) {
	input := `"hello": "world!"` // missing { }
	err := validateModuleArgs(input)
	require.NotNil(t, err)
}
