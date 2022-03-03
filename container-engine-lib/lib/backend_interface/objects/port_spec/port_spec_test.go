package port_spec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConstructorErrorsOnUnrecognizedProtocol(t *testing.T) {
	_, err := NewPortSpec(123, PortProtocol(999))
	require.Error(t, err)

}
