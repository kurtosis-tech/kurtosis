package partition_topology

import "fmt"

const percentageSuffix = "%"

var (
	ConnectionWithNoPacketLoss     = NewPacketLoss(0)
	ConnectionWithEntirePacketLoss = NewPacketLoss(100)
)

type PacketLoss struct {
	packetLossPercentage float32
}

func NewPacketLoss(packetLossPercentage float32) PacketLoss {
	return PacketLoss{
		packetLossPercentage: packetLossPercentage,
	}
}

// IsSet This method checks whether we need to set loss percentage, default value is 0
func (packetLoss *PacketLoss) IsSet() bool {
	return packetLoss.packetLossPercentage > 0
}

func (packetLoss *PacketLoss) GetTcCommand() string {
	packetLossMilliSecondStr := fmt.Sprintf("%v%v", packetLoss.packetLossPercentage, percentageSuffix)
	return packetLossMilliSecondStr
}
