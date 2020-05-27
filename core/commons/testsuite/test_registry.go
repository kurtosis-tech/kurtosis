package testsuite

import (
	"github.com/gmarchetti/kurtosis/commons/testnet"
	"github.com/palantir/stacktrace"
)

type TestRegistryBuilder struct {
	tests map[string]testnet.TestNetworkLoader
}

func NewTestRegistryBuilder() *TestRegistryBuilder {
	tests := make(map[string]testnet.TestNetworkLoader)
	return &TestRegistryBuilder{
		tests: tests,
	}
}

// Frustratingly, becuase of no generics, we can't guarantee that the network type returned by TestNetworkLoad and the
// network type consuemd by Test are the same - the user is going to have to cast it themselves inside the test :(
func (registry *TestRegistryBuilder) AddTest(name string, networkLoader testnet.TestNetworkLoader, test Test) error {
	if _, found := registry.tests[name]; found {
		return stacktrace.NewError("Test '%v' already registered", name)
	}
	registry.tests[name] = networkLoader
	return nil
}

type TestRegistry struct {
	tests map[string]testnet.TestNetworkLoader
}
