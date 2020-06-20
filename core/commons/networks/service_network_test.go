package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"os"
	"testing"
	"time"
)


type TestService struct {}

// ======================== Test Initializer Core ========================
type TestInitializerCore struct {}
func (t TestInitializerCore) GetUsedPorts() map[int]bool {
	return make(map[int]bool)
}

func (t TestInitializerCore) GetStartCommand(publicIpAddr string, dependencies []services.Service) ([]string, error) {
	return make([]string, 0), nil
}

func (t TestInitializerCore) GetServiceFromIp(ipAddr string) services.Service {
	return TestService{}
}


func (t TestInitializerCore) GetFilepathsToMount() map[string]bool {
	return make(map[string]bool)
}

func (t TestInitializerCore) InitializeMountedFiles(filepathsToMount map[string]*os.File, dependencies []services.Service) error {
	return nil
}

func getTestInitializerCore() services.ServiceInitializerCore {
	return TestInitializerCore{}
}


// ======================== Test Availability Checker Core ========================
type TestAvailabilityCheckerCore struct {}
func (t TestAvailabilityCheckerCore) IsServiceUp(toCheck services.Service, dependencies []services.Service) bool {
	return true
}
func (t TestAvailabilityCheckerCore) GetTimeout() time.Duration {
	return 30 * time.Second
}
func getTestCheckerCore() services.ServiceAvailabilityCheckerCore {
	return TestAvailabilityCheckerCore{}
}

// ======================== Tests ========================
func TestDisallowingNonexistentConfigs(t *testing.T) {
	builder := NewServiceNetworkBuilder("test", nil, nil)
	network := builder.Build()
	_, err := network.AddService(0, 0, make(map[int]bool))
	if err == nil {
		t.Fatal("Expected error when declaring a service with a configuration that doesn't exist")
	}
}

func TestDisallowingNonexistentDependencies(t *testing.T) {
	builder := NewServiceNetworkBuilder("test", nil, nil)
	configId := builder.AddTestImageConfiguration(getTestInitializerCore(), getTestCheckerCore())
	network := builder.Build()

	dependencies := map[int]bool{
		0: true,
	}

	_, err := network.AddService(configId, 0, dependencies)
	if err == nil {
		t.Fatal("Expected error when declaring a dependency on a service ID that doesn't exist")
	}
}
