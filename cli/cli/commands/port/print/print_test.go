package print

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testIpAddress       = "127.0.0.1"
	port                = 123
	applicationProtocol = "http"
	transportProtocol   = services.TransportProtocol_TCP
)

var portSpec = services.NewPortSpec(port, transportProtocol, applicationProtocol)

func TestOnlyPiece(t *testing.T) {
	inputExpectedOutputMap := map[string]string{
		"number":   fmt.Sprintf("%d", port),
		"protocol": applicationProtocol,
		"ip":       testIpAddress,
	}
	for in, expected := range inputExpectedOutputMap {
		out, err := formatPortOutput(in, testIpAddress, portSpec)
		assert.NoError(t, err)
		assert.Equal(t, expected, out)
	}
}

func TestAllPieces(t *testing.T) {
	out, err := formatPortOutput("protocol,ip,number", testIpAddress, portSpec)
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:123", out)
}

func TestPiecesOutOfOrder(t *testing.T) {
	_, err := formatPortOutput("ip,protocol", testIpAddress, portSpec)
	assert.Error(t, err)
}
