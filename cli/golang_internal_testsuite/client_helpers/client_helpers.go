package client_helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"time"
)

const (
	datastorePort                         = datastore_rpc_api_consts.ListenPort
	apiServicePort                        = example_api_server_rpc_api_consts.ListenPort
	configFilepathRelativeToSharedDirRoot = "config-file.txt"

	datastoreImage = "kurtosistech/example-datastore-server"
	apiServiceImage = "kurtosistech/example-api-server"
)

type GRPCAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort uint16    `json:"datastorePort"`
}

func GetDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func NewDatastoreClient(datastoreIp string) (datastore_rpc_api_bindings.DatastoreServiceClient, func() error, error) {
	datastoreURL := fmt.Sprintf(
		"%v:%v",
		datastoreIp,
		datastorePort,
	)

	conn, err := grpc.Dial(datastoreURL, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling the datastore container via its URL")
	}

	datastoreServiceClient := datastore_rpc_api_bindings.NewDatastoreServiceClient(conn)

	return datastoreServiceClient, conn.Close, nil
}

func WaitForHealthy(ctx context.Context, client GRPCAvailabilityChecker, retries uint32, retriesDelayMilliseconds uint32) error {

	var (
		emptyArgs = &empty.Empty{}
		err       error
	)

	for i := uint32(0); i < retries; i++ {
		_, err = client.IsAvailable(ctx, emptyArgs)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(err,
			"The datastore service didn't return a success code, even after %v retries with %v milliseconds in between retries",
			retries, retriesDelayMilliseconds)
	}

	return nil
}

func GetApiServiceContainerConfigSupplier(datastoreIP string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		datastoreConfigFileFilePath, err := CreateDatastoreConfigFileInServiceDirectory(datastoreIP, sharedDirectory)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating data store config file in service container")
		}

		startCmd := []string{
			"./example-api-server.bin",
			"--config",
			datastoreConfigFileFilePath.GetAbsPathOnServiceContainer(),
		}

		containerConfig := services.NewContainerConfigBuilder(
			apiServiceImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", apiServicePort): true},
		).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}

func CreateDatastoreConfigFileInServiceDirectory(datastoreIP string, sharedDirectory *services.SharedPath) (*services.SharedPath, error) {
	configFileFilePath := sharedDirectory.GetChildPath(configFilepathRelativeToSharedDirRoot)

	logrus.Infof("Config file absolute path on this container: %v , on service container: %v", configFileFilePath.GetAbsPathOnThisContainer(), configFileFilePath.GetAbsPathOnServiceContainer())

	logrus.Debugf("Datastore IP: %v , port: %v", datastoreIP, datastorePort)

	configObj := datastoreConfig{
		DatastoreIp:   datastoreIP,
		DatastorePort: datastorePort,
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	logrus.Debugf("API config JSON: %v", string(configBytes))

	if err := ioutil.WriteFile(configFileFilePath.GetAbsPathOnThisContainer(), configBytes, os.ModePerm); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return configFileFilePath, nil
}

func NewExampleAPIServerClient(exampleAPIServerIp string) (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {
	exampleAPIServerURL := fmt.Sprintf(
		"%v:%v",
		exampleAPIServerIp,
		apiServicePort,
	)

	conn, err := grpc.Dial(exampleAPIServerURL, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling the example API server container via its URL")
	}

	exampleAPIServerClient := example_api_server_rpc_api_bindings.NewExampleAPIServerServiceClient(conn)

	return exampleAPIServerClient, conn.Close, nil
}
