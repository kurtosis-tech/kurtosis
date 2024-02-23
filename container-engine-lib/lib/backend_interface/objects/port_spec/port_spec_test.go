package port_spec

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	httpAppProtocol = "https"
)

var portWaitForTest = NewWait(5 * time.Second)

func TestConstructorErrorsOnUnrecognizedProtocol(t *testing.T) {
	_, err := NewPortSpec(123, TransportProtocol(999), "", portWaitForTest, "")
	require.Error(t, err)
}

func TestNewPortSpec_WithApplicationProtocolPresent(t *testing.T) {
	spec, err := NewPortSpec(123, TransportProtocol_TCP, httpAppProtocol, portWaitForTest, "")

	privatePortSpec := &privatePortSpec{
		123,
		TransportProtocol_TCP,
		&httpAppProtocol,
		portWaitForTest,
		nil,
	}

	specActual := &PortSpec{
		privatePortSpec,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestNewPortSpec_WithApplicationProtocolAbsent(t *testing.T) {
	spec, err := NewPortSpec(123, TransportProtocol_TCP, "", portWaitForTest, "")

	privatePortSpec := &privatePortSpec{
		123,
		TransportProtocol_TCP,
		nil,
		portWaitForTest,
		nil,
	}

	specActual := &PortSpec{
		privatePortSpec,
	}

	require.NoError(t, err)
	require.Equal(t, spec, specActual)
}

func TestPortSpecMarshallers(t *testing.T) {
	originalPortSpec, err := NewPortSpec(123, TransportProtocol_TCP, httpAppProtocol, portWaitForTest, "")
	require.NoError(t, err)

	marshaledPortSpec, err := json.Marshal(originalPortSpec)
	require.NoError(t, err)
	require.NotNil(t, marshaledPortSpec)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newPortSpec := &PortSpec{}

	err = json.Unmarshal(marshaledPortSpec, newPortSpec)
	require.NoError(t, err)

	require.EqualValues(t, originalPortSpec, newPortSpec)
}
