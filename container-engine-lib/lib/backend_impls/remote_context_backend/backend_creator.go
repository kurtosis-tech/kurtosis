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
	urlScheme = "tcp"

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
	url := fmt.Sprintf("%s://%s:%d", urlScheme, remoteBackendConfig.Host, remoteBackendConfig.Port)
	clientOptions = append(clientOptions, client.WithHost(url))

	// TLS option if config is present
	if tlsConfig := remoteBackendConfig.Tls; tlsConfig != nil {
		tlsFilesDir, cleanCertFilesFunc, err := writeTlsConfigToTempDir(tlsConfig.Ca, tlsConfig.ClientCert, tlsConfig.ClientKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error building TLS configuration to connect to remote Docker backend")
		}
		defer cleanCertFilesFunc()
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
