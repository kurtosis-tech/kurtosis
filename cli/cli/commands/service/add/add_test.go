package add

import (
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsePortSpecstr_EmptyIsError(t *testing.T) {
	_, err := parsePortSpecStr("")
	require.Error(t, err)
}

func TestParsePortSpecstr_AlphabeticalIsError(t *testing.T) {
	_, err := parsePortSpecStr("abd")
	require.Error(t, err)
}

func TestParsePortSpecstr_TooManyDelimsIsError(t *testing.T) {
	_, err := parsePortSpecStr("1234/tcp/udp")
	require.Error(t, err)
}

func TestParsePortSpecstr_DefaultTcpProtocol(t *testing.T) {
	portSpec, err := parsePortSpecStr("1234")
	require.NoError(t, err)
	require.Equal(t, uint16(1234), portSpec.GetNumber())
	require.Equal(t, services.PortProtocol_TCP, portSpec.GetProtocol())
}

func TestParsePortSpecstr_CustomProtocol(t *testing.T) {
	portSpec, err := parsePortSpecStr("1234/udp")
	require.NoError(t, err)
	require.Equal(t, uint16(1234), portSpec.GetNumber())
	require.Equal(t, services.PortProtocol_UDP, portSpec.GetProtocol())
}
