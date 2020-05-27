package testsuite

import "github.com/gmarchetti/kurtosis/commons/testnet"

type TestConfig struct {
	Test Test
	NetworkLoader testnet.TestNetworkLoader
}
