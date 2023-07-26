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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"time"
)

const (
	dockerClientTimeout = 5 * time.Minute
)

// TODO Delete this when we split up KurtosisBackend into various parts
// Struct encapsulating information needed to prep the DockerKurtosisBackend for extended API container functionality
type APIContainerModeArgs struct {
	// Normally storing a context in a struct is bad, but we only do this to package it together as part of "optional" args
	Context        context.Context
	EnclaveID      enclave.EnclaveUUID
	APIContainerIP net.IP
}

// GetLocalDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
// ONLY the API container should pass in the extra API container args, which will unlock extra API container functionality
func GetLocalDockerKurtosisBackend(
	optionalApiContainerModeArgs *APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	localDockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithTimeout(dockerClientTimeout), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}

	localDockerBackend, err := GetDockerKurtosisBackend(localDockerClient, optionalApiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to build local Kurtosis Docker backend")
	}
	return localDockerBackend, nil
}

func GetDockerKurtosisBackend(
	dockerClient *client.Client,
	optionalApiContainerModeArgs *APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	dockerManager := docker_manager.NewDockerManager(dockerClient)

	// If running within the API container context, detect the network that the API container is running inside
	// so, we can create the free IP address trackers
	enclaveFreeIpAddrTrackers := map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker{}
	if optionalApiContainerModeArgs != nil {
		enclaveDb, err := enclave_db.GetOrCreateEnclaveDatabase()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred opening local database")
		}
		ctx := optionalApiContainerModeArgs.Context
		enclaveUuid := optionalApiContainerModeArgs.EnclaveID

		enclaveNetworkSearchLabels := map[string]string{
			label_key_consts.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
			label_key_consts.IDDockerLabelKey.GetString():    string(enclaveUuid),
		}
		matchingNetworks, err := dockerManager.GetNetworksByLabels(ctx, enclaveNetworkSearchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting Docker networks matching enclave ID '%v', which is necessary for the API "+
					"container to get the network CIDR and its own IP address",
				enclaveUuid,
			)
		}
		if len(matchingNetworks) == 0 {
			return nil, stacktrace.NewError("Didn't find any Docker networks matching enclave '%v'; this is a bug in Kurtosis", enclaveUuid)
		}
		if len(matchingNetworks) > 1 {
			return nil, stacktrace.NewError("Found more than one Docker network matching enclave '%v'; this is a bug in Kurtosis", enclaveUuid)
		}
		network := matchingNetworks[0]
		networkIp := network.GetIpAndMask().IP
		apiContainerIp := optionalApiContainerModeArgs.APIContainerIP

		alreadyTakenIps := map[string]bool{
			networkIp.String():      true,
			network.GetGatewayIp():  true,
			apiContainerIp.String(): true,
		}

		freeIpAddrProvider, err := free_ip_addr_tracker.GetOrCreateNewFreeIpAddrTracker(
			network.GetIpAndMask(),
			alreadyTakenIps,
			enclaveDb,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating IP tracker")
		}

		enclaveFreeIpAddrTrackers[enclaveUuid] = freeIpAddrProvider
	}

	dockerKurtosisBackend := docker_kurtosis_backend.NewDockerKurtosisBackend(dockerManager, enclaveFreeIpAddrTrackers)

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(dockerKurtosisBackend)

	return wrappedBackend, nil
}
