package fluentbit

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"

func createFluentbitConfigurationCreatorForKurtosis(
	enclaveUuid enclave.EnclaveUUID,
	logsDatabaseHost string,
	logsDatabasePort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *fluentbitConfigurationCreator {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(enclaveUuid, logsDatabaseHost, logsDatabasePort, tcpPortNumber, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitConfigurationCreator(config)
	return fluentbitContainerConfigProvider
}
