package remote_context_backend

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	urlProtocol = "tcp"

	noTempDirPrefix    = ""
	tempDirNamePattern = "kurtosis_backend_tls_*"
	caFileName         = "ca.pem"
	certFileName       = "cert.pem"
	keyFileName        = "key.pem"
	tlsFilesPerm       = 0644
)

func GetContextAwareKurtosisBackend(
	remoteBackendConfig *KurtosisRemoteBackendConfig,
	optionalApiContainerModeArgs *backend_creator.APIContainerModeArgs,
) (backend_interface.KurtosisBackend, error) {
	localDockerBackend, err := backend_creator.GetLocalDockerKurtosisBackend(optionalApiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to build Kurtosis Backend connected to local Docker instance")
	}

	if remoteBackendConfig == nil {
		logrus.Debugf("Instantiating a context aware backend with no remote backend config ends up returning" +
			"a regular local Docker backend.")
		return localDockerBackend, nil
	}
	remoteDockerClient, err := buildRemoteDockerClient(remoteBackendConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building client configuration for Docker remote backend")
	}
	kurtosisRemoteBackend, err := backend_creator.GetDockerKurtosisBackend(remoteDockerClient, optionalApiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building Kurtosis remote Docker backend")
	}
	return newRemoteContextKurtosisBackend(localDockerBackend, kurtosisRemoteBackend), nil
}

func buildRemoteDockerClient(remoteBackendConfig *KurtosisRemoteBackendConfig) (*client.Client, error) {
	var clientOptions []client.Opt

	// host and port option
	url := fmt.Sprintf("%s://%s:%d", urlProtocol, remoteBackendConfig.Host, remoteBackendConfig.Port)
	clientOptions = append(clientOptions, client.WithHost(url))

	// TLS option if config is present
	if tlsConfig := remoteBackendConfig.Tls; tlsConfig != nil {
		tlsFilesDir, err := writeTlsConfigToTempDir(tlsConfig.Ca, tlsConfig.ClientCert, tlsConfig.ClientKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error building TLS configuration to connect to remote Docker backend")
		}
		tlsOpt := client.WithTLSClientConfig(
			path.Join(tlsFilesDir, caFileName),
			path.Join(tlsFilesDir, certFileName),
			path.Join(tlsFilesDir, keyFileName))
		clientOptions = append(clientOptions, tlsOpt)
	}

	// API version negotiation option
	clientOptions = append(clientOptions, client.WithAPIVersionNegotiation())

	remoteDockerClient, err := client.NewClientWithOpts(clientOptions...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building Docker remote client")
	}
	return remoteDockerClient, nil
}

func writeTlsConfigToTempDir(ca []byte, cert []byte, key []byte) (string, error) {
	tempDirectory, err := os.MkdirTemp("", tempDirNamePattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Cannot create a temporary directory to store Kurtosis backend TLS files")
	}
	if err = os.WriteFile(path.Join(tempDirectory, caFileName), ca, tlsFilesPerm); err != nil {
		return "", stacktrace.Propagate(err, "Error writing content of CA to temporary file")
	}
	if err = os.WriteFile(path.Join(tempDirectory, certFileName), cert, tlsFilesPerm); err != nil {
		return "", stacktrace.Propagate(err, "Error writing content of certificate to temporary file")
	}
	if err = os.WriteFile(path.Join(tempDirectory, keyFileName), key, tlsFilesPerm); err != nil {
		return "", stacktrace.Propagate(err, "Error writing content of key to temporary file")
	}
	return tempDirectory, nil
}
