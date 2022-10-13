package kurtosis_types

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestPortSpec_StringRepresentation(t *testing.T) {
	portSpec := NewPortSpec(starlark.MakeInt(123), "TCP")
	expectedStr := "port_spec(number=123, protocol=\"TCP\")"
	require.Equal(t, expectedStr, portSpec.String())
}
