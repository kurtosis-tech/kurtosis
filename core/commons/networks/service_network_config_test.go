package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"gotest.tools/assert"
	"testing"
	"time"
)


type TestService struct {}

type TestFactoryConfig struct {}
func (t TestFactoryConfig) GetDockerImage() string {
	return "TEST"
}

func (t TestFactoryConfig) GetUsedPorts() map[int]bool {
	return make(map[int]bool)
}

func (t TestFactoryConfig) GetStartCommand(publicIpAddr string, dependencies []services.Service) []string {
	return make([]string, 0)
}

func (t TestFactoryConfig) GetServiceFromIp(ipAddr string) services.Service {
	return TestService{}
}

func (t TestFactoryConfig) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	return true
}

func (t TestFactoryConfig) GetStartupTimeout() time.Duration {
	return 30 * time.Second
}

func getTestServiceFactory() *services.ServiceFactory {
	return services.NewServiceFactory(TestFactoryConfig{})
}

func TestDisallowingNonexistentConfigs(t *testing.T) {
	builder := NewServiceNetworkConfigBuilder()
	_, err := builder.AddService(0, make(map[int]bool))
	if err == nil {
		t.Fatal("Expected error when declaring a service with a configuration that doesn't exist")
	}
}

func TestDisallowingNonexistentDependencies(t *testing.T) {
	builder := NewServiceNetworkConfigBuilder()
	config := builder.AddStaticImageConfiguration(*getTestServiceFactory())

	dependencies := map[int]bool{
		0: true,
	}

	_, err := builder.AddService(config, dependencies)
	if err == nil {
		t.Fatal("Expected error when declaring a dependency on a service ID that doesn't exist")
	}
}

// TODO test configuration IDs get incremented!

func TestIdsDifferent(t *testing.T) {
	builder := NewServiceNetworkConfigBuilder()
	config := builder.AddStaticImageConfiguration(*getTestServiceFactory())
	svc1, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	svc2, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	assert.Assert(t, svc1 != svc2, "IDs should be different")
}

func TestDependencyBookkeeping(t *testing.T) {
	builder := NewServiceNetworkConfigBuilder()
	config := builder.AddStaticImageConfiguration(*getTestServiceFactory())

	svc1, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc2, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc3Deps := map[int]bool{
		svc1: true,
		svc2: true,
	}
	svc3, err := builder.AddService(config, svc3Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc4Deps := map[int]bool{
		svc1: true,
		svc3: true,
	}
	svc4, err := builder.AddService(config, svc4Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc5Deps := map[int]bool{
		svc2: true,
	}
	svc5, err := builder.AddService(config, svc5Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}


	expectedOrder := []int{
		svc1,
		svc2,
		svc3,
		svc4,
		svc5,
	}
	assert.DeepEqual(t,
		expectedOrder,
		builder.servicesStartOrder)

	expectedDependents := map[int]bool{
		svc4: true,
		svc5: true,
	}
	if len(expectedDependents) != len(builder.onlyDependentServices) {
		t.Fatal("Size of dependent-only services didn't match expected")
	}
	for svcId := range builder.onlyDependentServices {
		if _, found := expectedDependents[svcId]; !found {
			t.Fatalf("ID %v should be marked as dependent-only, but wasn't", svcId)
		}
	}
}

func TestDefensiveCopies(t *testing.T) {
	builder := NewServiceNetworkConfigBuilder()
	config := builder.AddStaticImageConfiguration(*getTestServiceFactory())

	dependencyMap := make(map[int]bool)
	svc1, err := builder.AddService(config, dependencyMap)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	networkConfig := builder.Build()

	_ = builder.AddStaticImageConfiguration(*getTestServiceFactory())
	_, err = builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	assert.Equal(t, 1, len(networkConfig.onlyDependentServices))
	assert.Equal(t, 1, len(networkConfig.serviceConfigs))
	assert.Equal(t, 1, len(networkConfig.servicesStartOrder))
	assert.Equal(t, 1, len(networkConfig.configurations))

	svcDependencies := networkConfig.serviceDependencies
	assert.Equal(t, 1, len(svcDependencies))
	dependencyMap[99] = true
	assert.Equal(t, 0, len(svcDependencies[svc1]))

	// TODO test that the dependencies in the GetStartCommand are what we expect!
}


