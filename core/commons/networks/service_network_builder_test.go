package networks

import (
	"gotest.tools/v3/assert"
	"testing"
)

const (
	testConfigurationId0 = "test-configuration-0"
	testConfigurationId1 = "test-configuration-1"
)

func TestDisallowingSameIds(t *testing.T) {
	builder := NewServiceNetworkBuilder(nil, "test-network", nil, "test", "/foo/bar")
	err := builder.AddConfiguration(testConfigurationId0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}

	err = builder.AddConfiguration(testConfigurationId0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err == nil {
		t.Fatal("Expected an error here!")
	}
}

func TestDefensiveCopies(t *testing.T) {
	builder := NewServiceNetworkBuilder(nil, "test-network", nil, "test", "/foo/bar")
	err := builder.AddConfiguration(testConfigurationId0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}
	network := builder.Build()

	assert.Equal(t, 1, len(network.configurations))

	err = builder.AddConfiguration(testConfigurationId1, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}

	assert.Equal(t, 1, len(network.configurations))
}
