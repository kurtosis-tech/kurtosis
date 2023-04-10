package port_spec

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var portWaitForTest = NewWait(5 * time.Second)

func TestConstructorErrorsOnUnrecognizedProtocol(t *testing.T) {
	_, err := NewPortSpec(123, TransportProtocol(999), "", portWaitForTest)
	require.Error(t, err)
}

func TestNewPortSpec_WithApplicationProtocolPresent(t *testing.T) {
	https := "https"
	spec, err := NewPortSpec(123, TransportProtocol_TCP, https, portWaitForTest)

	specActual := &PortSpec{
		123,
		TransportProtocol_TCP,
		&https,
		portWaitForTest,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestNewPortSpec_WithApplicationProtocolAbsent(t *testing.T) {
	spec, err := NewPortSpec(123, TransportProtocol_TCP, "", portWaitForTest)

	specActual := &PortSpec{
		123,
		TransportProtocol_TCP,
		nil,
		portWaitForTest,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}
