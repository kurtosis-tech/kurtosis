package port_spec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConstructorErrorsOnUnrecognizedProtocol(t *testing.T) {
	_, err := NewPortSpec(123, TransportProtocol(999), "")
	require.Error(t, err)
}

func TestNewPortSpec_WithApplicationProtocolPresent(t *testing.T) {
	https := "https"
	spec, err := NewPortSpec(123, TransportProtocol_TCP, https)

	specActual := &PortSpec{
		123,
		TransportProtocol_TCP,
		&https,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestNewPortSpec_WithApplicationProtocolAbsent(t *testing.T) {
	spec, err := NewPortSpec(123, TransportProtocol_TCP, "")

	specActual := &PortSpec{
		123,
		TransportProtocol_TCP,
		nil,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}
