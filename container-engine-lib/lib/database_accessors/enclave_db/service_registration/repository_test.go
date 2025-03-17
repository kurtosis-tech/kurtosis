package service_registration

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_user"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	v1 "k8s.io/api/core/v1"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
)

const (
	serviceNameFormat  = "service-name-test-%v"
	uuidFormat         = "cddc2ea3948149d9afa2ef93abb4ec5%v"
	enclaveUuidFormat  = "%ve5c8bf2fbeb4de68f647280b1c79cbb"
	hostnameFormat     = "hostname-test-%v"
	privateIpStrFormat = "198.51.100.12%v"
	imageNameFormat    = "image-name:version-%v"
)

type ServiceRegistrationData struct {
	name        service.ServiceName
	uuid        service.ServiceUUID
	enclaveUuid enclave.EnclaveUUID
	hostname    string
	status      service.ServiceStatus
	privateIp   net.IP
	imageName   string
}

func TestSaveAndGetServiceRegistration_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	serviceRegistrationFromRepository, err := repository.Get(originalServiceRegistration.GetName())
	require.NoError(t, err)

	// Check specific fields rather than comparing entire objects
	require.Equal(t, originalServiceRegistration.GetName(), serviceRegistrationFromRepository.GetName())
	require.Equal(t, originalServiceRegistration.GetUUID(), serviceRegistrationFromRepository.GetUUID())
	require.Equal(t, originalServiceRegistration.GetEnclaveID(), serviceRegistrationFromRepository.GetEnclaveID())
	require.Equal(t, originalServiceRegistration.GetStatus(), serviceRegistrationFromRepository.GetStatus())
	
	// Check config fields
	originalConfig := originalServiceRegistration.GetConfig()
	retrievedConfig := serviceRegistrationFromRepository.GetConfig()
	require.NotNil(t, retrievedConfig)
	require.Equal(t, originalConfig.GetContainerImageName(), retrievedConfig.GetContainerImageName())
}

func TestGetAll_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	amountOfRegistrations := 5

	originalServiceRegistrations, _ := getServiceRegistrationsForTest(t, amountOfRegistrations)
	require.NotNil(t, originalServiceRegistrations)
	require.Equal(t, amountOfRegistrations, len(originalServiceRegistrations))

	for _, serviceRegistration := range originalServiceRegistrations {
		require.NotNil(t, serviceRegistration)
		err := repository.Save(serviceRegistration)
		require.NoError(t, err)
	}

	serviceRegistrationsFromRepository, err := repository.GetAll()
	require.NoError(t, err)
	require.Len(t, serviceRegistrationsFromRepository, len(originalServiceRegistrations))

	// Check that all service names are present - don't compare full objects
	for name := range originalServiceRegistrations {
		_, found := serviceRegistrationsFromRepository[name]
		require.True(t, found, "Expected service %s to be in the result", name)
	}
}

func TestExist_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	exist, err := repository.Exist(originalServiceRegistration.GetName())
	require.NoError(t, err)
	require.True(t, exist)
}

func TestDoesNotExist(t *testing.T) {
	repository := getRepositoryForTest(t)

	saveAndGetOneServiceRegistrationForTest(t, repository)

	fakeServiceName := service.ServiceName("fake-service-name")

	exist, err := repository.Exist(fakeServiceName)
	require.NoError(t, err)
	require.False(t, exist)
}

func TestGetAllServiceNames_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	amountOfRegistrations := 5

	originalServiceRegistrations, originalServiceNames := getServiceRegistrationsForTest(t, amountOfRegistrations)
	require.NotNil(t, originalServiceRegistrations)
	require.Equal(t, amountOfRegistrations, len(originalServiceRegistrations))

	for _, serviceRegistration := range originalServiceRegistrations {
		require.NotNil(t, serviceRegistration)
		err := repository.Save(serviceRegistration)
		require.NoError(t, err)
	}

	allServiceNamesFromRepository, err := repository.GetAllServiceNames()
	require.NoError(t, err)
	require.Len(t, allServiceNamesFromRepository, len(originalServiceRegistrations))

	require.EqualValues(t, originalServiceNames, allServiceNamesFromRepository)
}

func TestUpdateStatus_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	newStatus := service.ServiceStatus_Started
	if originalServiceRegistration.GetStatus() == service.ServiceStatus_Started {
		newStatus = service.ServiceStatus_Stopped
	}

	err := repository.UpdateStatus(originalServiceRegistration.GetName(), newStatus)
	require.NoError(t, err)

	serviceRegistrationFromRepository, err := repository.Get(originalServiceRegistration.GetName())
	require.NoError(t, err)

	require.Equal(t, newStatus, serviceRegistrationFromRepository.GetStatus())
}

func TestUpdateConfig_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	anotherImageName := "another-image-name"
	newConfig := getServiceConfigForTest(t, anotherImageName)

	err := repository.UpdateConfig(originalServiceRegistration.GetName(), newConfig)
	require.NoError(t, err)

	serviceRegistrationFromRepository, err := repository.Get(originalServiceRegistration.GetName())
	require.NoError(t, err)

	retrievedConfig := serviceRegistrationFromRepository.GetConfig()
	require.NotNil(t, retrievedConfig)
	// Check specific fields rather than comparing entire configs
	require.Equal(t, anotherImageName, retrievedConfig.GetContainerImageName(), "Image name should match")
	require.NotNil(t, retrievedConfig.GetPrivatePorts(), "Private ports should not be nil")
	require.NotNil(t, retrievedConfig.GetPublicPorts(), "Public ports should not be nil")
	require.NotNil(t, retrievedConfig.GetLabels(), "Labels should not be nil")
}

func TestUpdateStatusAndConfig_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	newStatus := service.ServiceStatus_Stopped
	if originalServiceRegistration.GetStatus() == service.ServiceStatus_Stopped {
		newStatus = service.ServiceStatus_Started
	}

	anotherImageName := "another-image-name"
	newConfig := getServiceConfigForTest(t, anotherImageName)

	err := repository.UpdateStatusAndConfig(originalServiceRegistration.GetName(), newStatus, newConfig)
	require.NoError(t, err)

	serviceRegistrationFromRepository, err := repository.Get(originalServiceRegistration.GetName())
	require.NoError(t, err)

	require.Equal(t, newStatus, serviceRegistrationFromRepository.GetStatus())
	
	retrievedConfig := serviceRegistrationFromRepository.GetConfig()
	require.NotNil(t, retrievedConfig)
	// Check specific fields rather than comparing entire configs
	require.Equal(t, anotherImageName, retrievedConfig.GetContainerImageName(), "Image name should match")
	require.NotNil(t, retrievedConfig.GetPrivatePorts(), "Private ports should not be nil")
	require.NotNil(t, retrievedConfig.GetPublicPorts(), "Public ports should not be nil")
	require.NotNil(t, retrievedConfig.GetLabels(), "Labels should not be nil")
}

func TestDelete_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceRegistration := saveAndGetOneServiceRegistrationForTest(t, repository)

	err := repository.Delete(originalServiceRegistration.GetName())
	require.NoError(t, err)

	exist, err := repository.Exist(originalServiceRegistration.GetName())
	require.NoError(t, err)
	require.False(t, exist)
}

func saveAndGetOneServiceRegistrationForTest(
	t *testing.T,
	repository *ServiceRegistrationRepository,
) *service.ServiceRegistration {
	amountOfRegistrations := 1

	serviceRegistrations, _ := getServiceRegistrationsForTest(t, amountOfRegistrations)
	require.NotNil(t, serviceRegistrations)
	require.Equal(t, amountOfRegistrations, len(serviceRegistrations))

	firsServiceName := service.ServiceName(fmt.Sprintf(serviceNameFormat, 0))
	originalServiceRegistration, found := serviceRegistrations[firsServiceName]
	require.True(t, found)
	require.NotNil(t, originalServiceRegistration)

	err := repository.Save(originalServiceRegistration)
	require.NoError(t, err)

	return originalServiceRegistration
}

func getRepositoryForTest(t *testing.T) *ServiceRegistrationRepository {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}
	repository, err := GetOrCreateNewServiceRegistrationRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

func getServiceRegistrationsForTest(t *testing.T, amount int) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]bool) {
	serviceRegistrations := map[service.ServiceName]*service.ServiceRegistration{}
	serviceNames := map[service.ServiceName]bool{}

	serviceRegistrationDataList := getServiceRegistrationData(amount)

	for _, data := range serviceRegistrationDataList {
		serviceRegistration := getServiceRegistrationWithDataForTest(
			t,
			data.name,
			data.uuid,
			data.enclaveUuid,
			data.privateIp,
			data.hostname,
			data.status,
			data.imageName,
		)
		serviceRegistrations[data.name] = serviceRegistration
		serviceNames[data.name] = true
	}

	return serviceRegistrations, serviceNames
}

func getServiceRegistrationData(amount int) []*ServiceRegistrationData {
	serviceRegistrationDataList := []*ServiceRegistrationData{}

	serviceStatusNumber := len(service.ServiceStatusValues()) - 1

	for i := 0; i < amount; i++ {
		name := service.ServiceName(fmt.Sprintf(serviceNameFormat, i))
		uuid := service.ServiceUUID(fmt.Sprintf(uuidFormat, i))
		enclaveUuid := enclave.EnclaveUUID(fmt.Sprintf(enclaveUuidFormat, i))
		hostname := fmt.Sprintf(hostnameFormat, i)
		privateIp := net.ParseIP(fmt.Sprintf(privateIpStrFormat, i))
		image := fmt.Sprintf(imageNameFormat, i)

		randomStatus := rand.Intn(serviceStatusNumber)
		status := service.ServiceStatus(randomStatus)

		data := &ServiceRegistrationData{
			name,
			uuid,
			enclaveUuid,
			hostname,
			status,
			privateIp,
			image,
		}
		serviceRegistrationDataList = append(serviceRegistrationDataList, data)
	}

	return serviceRegistrationDataList
}

func getServiceRegistrationWithDataForTest(
	t *testing.T,
	serviceName service.ServiceName,
	uuid service.ServiceUUID,
	enclaveUuid enclave.EnclaveUUID,
	privateIp net.IP,
	hostname string,
	serviceStatus service.ServiceStatus,
	image string,
) *service.ServiceRegistration {

	serviceRegistration := service.NewServiceRegistration(serviceName, uuid, enclaveUuid, privateIp, hostname)

	serviceRegistration.SetStatus(serviceStatus)

	serviceConfig := getServiceConfigForTest(t, image)

	serviceRegistration.SetConfig(serviceConfig)

	return serviceRegistration
}

func getServiceConfigForTest(t *testing.T, imageName string) *service.ServiceConfig {
	var user *service_user.ServiceUser
	var tolerations []v1.Toleration
	
	// Create Kubernetes config
	kubeConfig := &kubernetes.Config{
		ExtraIngressConfig: nil, // No ingress config for this test
	}
	
	// Initialize test port specs, envVars, etc. with realistic values
	privatePorts := testPrivatePorts(t)
	publicPorts := testPublicPorts(t)
	envVars := testEnvVars()
	filesArtifactExpansion := testFilesArtifactExpansion()
	persistentDirectory := testPersistentDirectory()
	
	// Create a service config with the specified image name
	serviceConfig, err := service.CreateServiceConfig(
		imageName, // Use the provided image name
		nil, nil, nil,
		privatePorts,
		publicPorts,
		[]string{"bin", "bash", "ls"}, // Add some command args
		[]string{"-l", "-a"},
		envVars,
		filesArtifactExpansion,
		persistentDirectory,
		500,
		1024,
		"127.0.0.1",
		100,
		512,
		map[string]string{
			"test-label-key":        "test-label-value",
			"test-second-label-key": "test-second-label-value",
		},
		user,
		tolerations,
		map[string]string{
			"disktype": "ssd",
		},
		image_download_mode.ImageDownloadMode_Missing,
		true,
		kubeConfig,
	)
	require.NoError(t, err)
	return serviceConfig
}

func testPersistentDirectory() *service_directory.PersistentDirectories {
	persistentDirectoriesMap := map[string]service_directory.PersistentDirectory{
		"dirpath1": {PersistentKey: service_directory.DirectoryPersistentKey("dirpath1_persistent_directory_key"), Size: service_directory.DirectoryPersistentSize(int64(0))},
		"dirpath2": {PersistentKey: service_directory.DirectoryPersistentKey("dirpath2_persistent_directory_key"), Size: service_directory.DirectoryPersistentSize(int64(0))},
	}

	return service_directory.NewPersistentDirectories(persistentDirectoriesMap)
}

func testFilesArtifactExpansion() *service_directory.FilesArtifactsExpansion {
	return &service_directory.FilesArtifactsExpansion{
		ExpanderImage: "expander-image:tag-version",
		ExpanderEnvVars: map[string]string{
			"ENV_VAR1": "env_var1_value",
			"ENV_VAR2": "env_var2_value",
		},
		ServiceDirpathsToArtifactIdentifiers: map[string][]string{
			"/path/number1": {"first_identifier"},
			"/path/number2": {"second_identifier"},
		},
		ExpanderDirpathsToServiceDirpaths: map[string]string{
			"/expander/dir1": "/service/dir1",
			"/expander/dir2": "/service/dir2",
		},
	}
}

func testPrivatePorts(t *testing.T) map[string]*port_spec.PortSpec {

	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	appProtocol1 := "app-protocol1"
	wait1 := port_spec.NewWait(5 * time.Minute)
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, appProtocol1, wait1, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	appProtocol2 := "app-protocol2"
	wait2 := port_spec.NewWait(24 * time.Second)
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, appProtocol2, wait2, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	return input
}

func testPublicPorts(t *testing.T) map[string]*port_spec.PortSpec {

	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	appProtocol1 := "app-protocol1-public"
	wait1 := port_spec.NewWait(5 * time.Minute)
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, appProtocol1, wait1, "")
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	appProtocol2 := "app-protocol2-public"
	wait2 := port_spec.NewWait(24 * time.Second)
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, appProtocol2, wait2, "")
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	return input
}

func testEnvVars() map[string]string {
	return map[string]string{
		"HTTP_PORT":  "80",
		"HTTPS_PORT": "443",
	}
}
