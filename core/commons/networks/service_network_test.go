package networks

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"os"
	"testing"
	"time"
)

const (
	TEST_SERVICE = "test-service"
	TEST_NETWORK = "test-network"
)

type TestService struct {}

// ======================== Test Initializer Core ========================
type TestInitializerCore struct {}
func (t TestInitializerCore) GetUsedPorts() map[nat.Port]bool {
	return make(map[nat.Port]bool)
}

func (t TestInitializerCore) GetServiceFromIp(ipAddr string) services.Service {
	return TestService{}
}


func (t TestInitializerCore) GetFilesToMount() map[string]bool {
	return make(map[string]bool)
}

func (t TestInitializerCore) InitializeMountedFiles(filepathsToMount map[string]*os.File, dependencies []services.Service) error {
	return nil
}

func (t TestInitializerCore) GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr string, dependencies []services.Service) ([]string, error) {
	return make([]string, 0), nil
}

func (t TestInitializerCore) GetTestVolumeMountpoint() string {
	return "/foo/bar"
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
	builder := NewServiceNetworkBuilder(nil, TEST_NETWORK, nil, "test", "/foo/bar")
	network := builder.Build()
	_, err := network.AddService(0, TEST_SERVICE, make(map[ServiceID]bool))
	if err == nil {
		t.Fatal("Expected error when declaring a service with a configuration that doesn't exist")
	}
}

func TestDisallowingNonexistentDependencies(t *testing.T) {
	var configId ConfigurationID = 0
	builder := NewServiceNetworkBuilder(nil, TEST_NETWORK, nil, "test", "/foo/bar")
	err := builder.AddConfiguration(configId, "test", getTestInitializerCore(), getTestCheckerCore())
	if err != nil {
		t.Fatal("Adding a configuration shouldn't fail")
	}
	network := builder.Build()

	dependencies := map[ServiceID]bool{
		TEST_SERVICE: true,
	}

	_, err = network.AddService(configId, TEST_SERVICE, dependencies)
	if err == nil {
		t.Fatal("Expected error when declaring a dependency on a service ID that doesn't exist")
	}
}
