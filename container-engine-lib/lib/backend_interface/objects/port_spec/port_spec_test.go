package port_spec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConstructorErrorsOnUnrecognizedProtocol(t *testing.T) {
	_, err := NewPortSpec(123, PortProtocol(999))
	require.Error(t, err)
}

func TestNewPortSpecWithApplicationProtocol(t *testing.T) {
	spec, err := NewPortSpec(123, PortProtocol_TCP, HTTPS)
	httpPtr := HTTPS

	specActual := &PortSpec{
		123,
		PortProtocol_TCP,
		&httpPtr,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestNewPortSpecWithoutApplicationProtocol(t *testing.T) {
	spec, err := NewPortSpec(123, PortProtocol_TCP)

	specActual := &PortSpec{
		123,
		PortProtocol_TCP,
		nil,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestNewPortSpecError(t *testing.T) {
	_, err := NewPortSpec(123, PortProtocol_TCP, HTTPS, HTTP)
	require.ErrorContains(t, err, "Application Protocol can have at most 1 value")
}
