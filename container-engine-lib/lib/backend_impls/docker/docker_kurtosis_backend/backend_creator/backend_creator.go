package backend_creator

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/metrics_reporting"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/struct_persister"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"net"
)

// TODO Delete this when we split up KurtosisBackend into various parts
// Struct encapsulating information needed to prep the DockerKurtosisBackend for extended API container functionality
type APIContainerModeArgs struct {
	// Normally storing a context in a struct is bad, but we only do this to package it together as part of "optional" args
	Context        context.Context
	EnclaveID      enclave.EnclaveID
	APIContainerIP net.IP
}

const (
	databaseFilePath              = "kurtosis.db"
	readWritePermissionToDatabase = 0666
)

// GetLocalDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
// ONLY the API container should pass in the extra API container args, which will unlock extra API container functionality
func GetLocalDockerKurtosisBackend(
	optionalApiContainerModeArgs *APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	db, err := bolt.Open(databaseFilePath, readWritePermissionToDatabase, &bolt.Options{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening local database")
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}

	dockerManager := docker_manager.NewDockerManager(dockerClient)

	// If running within the API container context, detect the network that the API container is running inside
	// so we can create the free IP address trackers
	enclaveFreeIpAddrTrackers := map[enclave.EnclaveID]*struct_persister.FreeIpAddrTracker{}
	if optionalApiContainerModeArgs != nil {
		ctx := optionalApiContainerModeArgs.Context
		enclaveId := optionalApiContainerModeArgs.EnclaveID

		enclaveNetworkSearchLabels := map[string]string{
			label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
			label_key_consts.IDDockerLabelKey.GetString():    string(enclaveId),
		}
		matchingNetworks, err := dockerManager.GetNetworksByLabels(ctx, enclaveNetworkSearchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting Docker networks matching enclave ID '%v', which is necessary for the API "+
					"container to get the network CIDR and its own IP address",
				enclaveId,
			)
		}
		if len(matchingNetworks) == 0 {
			return nil, stacktrace.NewError("Didn't find any Docker networks matching enclave '%v'; this is a bug in Kurtosis", enclaveId)
		}
		if len(matchingNetworks) > 1 {
			return nil, stacktrace.NewError("Found more than one Docker network matching enclave '%v'; this is a bug in Kurtosis", enclaveId)
		}
		network := matchingNetworks[0]
		networkIp := network.GetIpAndMask().IP
		apiContainerIp := optionalApiContainerModeArgs.APIContainerIP

		alreadyTakenIps := map[string]bool{
			networkIp.String():      true,
			network.GetGatewayIp():  true,
			apiContainerIp.String(): true,
		}

		freeIpAddrProvider, err := struct_persister.GetOrCreateNewFreeIpAddrTracker(
			logrus.StandardLogger(),
			network.GetIpAndMask(),
			alreadyTakenIps,
			db,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating IP tracker")
		}

		enclaveFreeIpAddrTrackers[enclaveId] = freeIpAddrProvider
	}

	dockerKurtosisBackend := docker_kurtosis_backend.NewDockerKurtosisBackend(dockerManager, enclaveFreeIpAddrTrackers)

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(dockerKurtosisBackend)

	return wrappedBackend, nil
}
