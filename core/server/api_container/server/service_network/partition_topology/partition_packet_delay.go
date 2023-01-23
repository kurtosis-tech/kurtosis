package partition_topology

import "fmt"

const milliSecondSuffix = "ms"

var (
	ConnectionWithNoPacketDelay = NewPacketDelay(0)
)

// PacketDelay - uses to en capture delay parameters the https://man7.org/linux/man-pages/man8/tc-netem.8.html
// We will only support NORMAL distribution, but others can be added such as pareto
// We can introduce distribution as an enum
// TODO: allow users to set jitter and correlation values via starlark
type PacketDelay struct {
	avgDelayMs uint32
}

func NewPacketDelay(packetDelayInMs uint32) PacketDelay {
	return PacketDelay{
		avgDelayMs: packetDelayInMs,
	}
}

// IsSet This method checks whether we need to set delay, default value is 0
func (packetDelay *PacketDelay) IsSet() bool {
	return packetDelay.avgDelayMs != 0
}

func (packetDelay *PacketDelay) GetTcCommand() string {
	packetDelayMilliSecondStr := fmt.Sprintf("%v%v", packetDelay.avgDelayMs, milliSecondSuffix)
	return packetDelayMilliSecondStr
}
