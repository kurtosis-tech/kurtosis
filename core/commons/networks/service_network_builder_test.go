package networks

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestDisallowingSameIds(t *testing.T) {
	builder := NewServiceNetworkBuilder(nil, "test-network", nil, "test", "/foo/bar")
	err := builder.AddConfiguration(0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}

	err = builder.AddConfiguration(0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err == nil {
		t.Fatal("Expected an error here!")
	}
}

func TestDefensiveCopies(t *testing.T) {
	builder := NewServiceNetworkBuilder(nil, "test-network", nil, "test", "/foo/bar")
	err := builder.AddConfiguration(0, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}
	network := builder.Build()

	assert.Equal(t, 1, len(network.configurations))

	err = builder.AddConfiguration(1, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail here")
	}

	assert.Equal(t, 1, len(network.configurations))
}
