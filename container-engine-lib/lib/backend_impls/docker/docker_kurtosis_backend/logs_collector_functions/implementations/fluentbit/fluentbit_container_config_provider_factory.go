package fluentbit

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"

func createFluentbitContainerConfigProviderForKurtosis(
	enclaveUuid enclave.EnclaveUUID,
	logsDatabaseHost string,
	logsDatabasePort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *fluentbitContainerConfigProvider {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(enclaveUuid, logsDatabaseHost, logsDatabasePort, tcpPortNumber, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitContainerConfigProvider(config, tcpPortNumber, httpPortNumber)
	return fluentbitContainerConfigProvider
}
