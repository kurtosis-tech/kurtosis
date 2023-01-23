package partition_topology

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPacketDelay_GetTcCommand(t *testing.T) {
	packetDelay := NewPacketDelay(100)
	actualTcCommand := packetDelay.GetTcCommand()
	expectedTcCommand := "100ms"
	require.Equal(t, expectedTcCommand, actualTcCommand)
}
