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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/configs"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
)

const (
	noTempDirPrefix    = ""
	tempDirNamePattern = "kurtosis_backend_tls_*"
	caFileName         = "ca.pem"
	certFileName       = "cert.pem"
	keyFileName        = "key.pem"
	tlsFilesPerm       = 0644
)

// TODO Delete this when we split up KurtosisBackend into various parts
// Struct encapsulating information needed to prep the DockerKurtosisBackend for extended API container functionality
type APIContainerModeArgs struct {
	// Normally storing a context in a struct is bad, but we only do this to package it together as part of "optional" args
	Context        context.Context
	EnclaveID      enclave.EnclaveUUID
	APIContainerIP net.IP
}

var (
	NoAPIContainerModeArgs *APIContainerModeArgs = nil
)

// GetDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
// It creates a local or remote docker backend based on the existence of a remote backend config.
// ONLY the API container should pass in the extra API container args, which will unlock extra API container functionality
func GetDockerKurtosisBackend(
	optionalApiContainerModeArgs *APIContainerModeArgs,
	optionalRemoteBackendConfig *configs.KurtosisRemoteBackendConfig,
) (backend_interface.KurtosisBackend, error) {
	var kurtosisBackend backend_interface.KurtosisBackend
	var err error
	if optionalRemoteBackendConfig != nil {
		kurtosisBackend, err = getRemoteDockerKurtosisBackend(optionalApiContainerModeArgs, optionalRemoteBackendConfig)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a remote Docker backend")
		}
	} else {
		kurtosisBackend, err = getLocalDockerKurtosisBackend(optionalApiContainerModeArgs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a local Docker backend")
		}
	}
	return kurtosisBackend, nil
}

// getLocalDockerKurtosisBackend is a Docker backend running locally
func getLocalDockerKurtosisBackend(
	optionalApiContainerModeArgs *APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	dockerClientOpts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}

	localDockerBackend, err := getDockerKurtosisBackend(dockerClientOpts, optionalApiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to build local Kurtosis Docker backend")
	}
	return localDockerBackend, nil
}

// getRemoteDockerKurtosisBackend is a Docker backend running on a remote host
func getRemoteDockerKurtosisBackend(
	optionalApiContainerModeArgs *APIContainerModeArgs,
	remoteBackendConfig *configs.KurtosisRemoteBackendConfig,
) (backend_interface.KurtosisBackend, error) {
	remoteDockerClientOpts, cleanCertFilesFunc, err := buildRemoteDockerClientOpts(remoteBackendConfig)
	if cleanCertFilesFunc != nil {
		defer cleanCertFilesFunc()
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building client configuration for Docker remote backend")
	}
	kurtosisRemoteBackend, err := getDockerKurtosisBackend(remoteDockerClientOpts, optionalApiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building Kurtosis remote Docker backend")
	}
	return kurtosisRemoteBackend, nil
}

func buildRemoteDockerClientOpts(remoteBackendConfig *configs.KurtosisRemoteBackendConfig) ([]client.Opt, func(), error) {
	var clientOptions []client.Opt

	// host and port option
	clientOptions = append(clientOptions, client.WithHost(remoteBackendConfig.Endpoint))

	// TLS option if config is present
	var tlsFilesDir string
	var cleanCertFilesFunc func()
	var err error
	if tlsConfig := remoteBackendConfig.Tls; tlsConfig != nil {
		tlsFilesDir, cleanCertFilesFunc, err = writeTlsConfigToTempDir(tlsConfig.Ca, tlsConfig.ClientCert, tlsConfig.ClientKey)
		if err != nil {
			return nil, cleanCertFilesFunc, stacktrace.Propagate(err, "Error building TLS configuration to connect to remote Docker backend")
		}
		tlsOpt := client.WithTLSClientConfig(
			path.Join(tlsFilesDir, caFileName),
			path.Join(tlsFilesDir, certFileName),
			path.Join(tlsFilesDir, keyFileName))
		clientOptions = append(clientOptions, tlsOpt)
	}

	// Timeout and API version negotiation option
	clientOptions = append(clientOptions, client.WithAPIVersionNegotiation())
	return clientOptions, cleanCertFilesFunc, nil
}

// writeTlsConfigToTempDir writes the different TLS files to a directory, and returns the path to this directory.
// It also returns a function to manually delete those files once they've been used upstream
func writeTlsConfigToTempDir(ca []byte, cert []byte, key []byte) (string, func(), error) {
	tempDirectory, err := os.MkdirTemp(noTempDirPrefix, tempDirNamePattern)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Cannot create a temporary directory to store Kurtosis backend TLS files")
	}
	caAbsFileName := path.Join(tempDirectory, caFileName)
	if err = os.WriteFile(caAbsFileName, ca, tlsFilesPerm); err != nil {
		return "", nil, stacktrace.Propagate(err, "Error writing content of CA to temporary file at '%s'", caAbsFileName)
	}
	certAbsFileName := path.Join(tempDirectory, certFileName)
	if err = os.WriteFile(certAbsFileName, cert, tlsFilesPerm); err != nil {
		return "", nil, stacktrace.Propagate(err, "Error writing content of certificate to temporary file at '%s'", certAbsFileName)
	}
	keyAbsFileName := path.Join(tempDirectory, keyFileName)
	if err = os.WriteFile(keyAbsFileName, key, tlsFilesPerm); err != nil {
		return "", nil, stacktrace.Propagate(err, "Error writing content of key to temporary file at '%s'", keyAbsFileName)
	}

	cleanDirectoryFunc := func() {
		if err = os.RemoveAll(tempDirectory); err != nil {
			logrus.Warnf("Error removing TLS config directory at '%s'. Will remain in the OS temporary files folder until the OS removes it", tempDirectory)
		}
	}
	return tempDirectory, cleanDirectoryFunc, nil
}

func getDockerKurtosisBackend(
	dockerClientOpts []client.Opt,
	optionalApiContainerModeArgs *APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	dockerManager, err := docker_manager.CreateDockerManager(dockerClientOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred building docker manager")
	}

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
