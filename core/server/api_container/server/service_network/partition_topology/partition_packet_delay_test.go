package partition_topology

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPacketDelay_GetTcCommand(t *testing.T) {
	packetDelay := NewUniformPacketDelayDistribution(100)
	actualTcCommand := packetDelay.GetTcCommand()
	expectedTcCommand := "100ms 0ms 0%"
	require.Equal(t, expectedTcCommand, actualTcCommand)

	packetDelay = NewNormalPacketDelayDistribution(100, 20, 20.5)
	actualTcCommand = packetDelay.GetTcCommand()
	expectedTcCommand = "100ms 20ms 20.5%"
	require.Equal(t, expectedTcCommand, actualTcCommand)
}
