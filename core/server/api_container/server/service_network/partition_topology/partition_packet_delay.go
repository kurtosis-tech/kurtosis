package partition_topology

import "fmt"

const milliSecondSuffix = "ms"

var (
	ConnectionWithNoPacketDelay = NewUniformPacketDelayDistribution(0)
)

// PacketDelayDistribution - uses to en capture delay parameters the https://man7.org/linux/man-pages/man8/tc-netem.8.html
// We will only support NORMAL distribution, but others can be added such as pareto
// We can introduce distribution as an enum
// TODO: allow users to set jitter and correlation values via starlark
type PacketDelayDistribution struct {
	avgDelayMs  uint32
	jitter      uint32
	correlation float32
}

func NewUniformPacketDelayDistribution(avgDelayMs uint32) PacketDelayDistribution {
	return PacketDelayDistribution{
		avgDelayMs:  avgDelayMs,
		jitter:      0,
		correlation: 0,
	}
}

func NewNormalPacketDelayDistribution(avgDelayMs uint32, jitter uint32, correlation float32) PacketDelayDistribution {
	return PacketDelayDistribution{
		avgDelayMs:  avgDelayMs,
		correlation: correlation,
		jitter:      jitter,
	}
}

// IsSet This method checks whether we require to set packet delay using tc command
func (packetDelay *PacketDelayDistribution) IsSet() bool {
	return packetDelay.avgDelayMs != 0 || packetDelay.jitter != 0 || packetDelay.correlation != 0
}

func (packetDelay *PacketDelayDistribution) GetTcCommand() string {
	packetCorrelationPercentage := fmt.Sprintf("%v%v", packetDelay.correlation, percentageSuffix)
	packetDelayMilliSecondStr := fmt.Sprintf("%v%v", packetDelay.avgDelayMs, milliSecondSuffix)
	packetJitterMilliSecondStr := fmt.Sprintf("%v%v", packetDelay.jitter, milliSecondSuffix)
	return fmt.Sprintf("%v %v %v", packetDelayMilliSecondStr, packetJitterMilliSecondStr, packetCorrelationPercentage)
}
