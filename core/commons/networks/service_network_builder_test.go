package networks

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestConfigurationIdsDifferent(t *testing.T) {
	testImage := "testImage"
	idSet := make(map[int]bool)
	builder := NewServiceNetworkBuilder("test", nil, nil)
	config1 := builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())
	idSet[config1] = true
	config2 := builder.AddStaticImageConfiguration(testImage, getTestInitializerCore(), getTestCheckerCore())
	idSet[config2] = true
	config3 := builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())
	idSet[config3] = true
	config4 := builder.AddStaticImageConfiguration(testImage, getTestInitializerCore(), getTestCheckerCore())
	idSet[config4] = true
	config5 := builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())
	idSet[config5] = true
	assert.Assert(t, len(idSet) == 5, "IDs should be different.")
}
func TestDefensiveCopies(t *testing.T) {
	builder := NewServiceNetworkBuilder("test", nil, nil)
	_ = builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())
	network := builder.Build()

	assert.Equal(t, 1, len(network.configurations))

	_ = builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())

	assert.Equal(t, 1, len(network.configurations))
}
