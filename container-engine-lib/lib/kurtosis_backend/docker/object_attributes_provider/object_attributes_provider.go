package schema

import (
	"fmt"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
	"time"
)

const (
	engineServerNamePrefix                   = "kurtosis-engine"
	engineServerPortProtocol                 = PortProtocol_TCP
	metadataAcquisitionContainerNameFragment = "metadata-acquisition"
)

type ObjectAttributesProvider interface {
	ForEngineServer(grpcListenPortNum uint16, grpcProxyListenPortNum uint16) (ObjectAttributes, error)
	ForEnclave(enclaveId string) EnclaveObjectAttributesProvider
}

// Entrypoint to get this version of the schema
func GetObjectAttributesProvider() ObjectAttributesProvider {
	return newObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type objectAttributesProviderImpl struct{}

func newObjectAttributesProviderImpl() *objectAttributesProviderImpl {
	return &objectAttributesProviderImpl{}
}

func (provider *objectAttributesProviderImpl) ForEngineServer(grpcListenPortNum uint16, grpcProxyListenPortNum uint16) (ObjectAttributes, error) {
	containerStartTimeUnixSecs := time.Now().Unix()
	containerStartTimeStr := fmt.Sprintf("%v", containerStartTimeUnixSecs)
	name := strings.Join(
		[]string{
			engineServerNamePrefix,
			containerStartTimeStr,
		},
		objectNameElementSeparator,
	)

	grpcPortSpec, err := NewPortSpec(grpcListenPortNum, engineServerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating grpc port spec object from num '%v' and protocol '%v'",
			grpcListenPortNum,
			engineServerPortProtocol,
		)
	}

	grpcProxyPortSpec, err := NewPortSpec(grpcProxyListenPortNum, engineServerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating grpc-proxy port spec object from num '%v' and protocol '%v'",
			grpcProxyListenPortNum,
			engineServerPortProtocol,
		)
	}

	usedPorts := map[string]*PortSpec{
		KurtosisInternalContainerGRPCPortID: grpcPortSpec,
		KurtosisInternalContainerGRPCProxyPortID: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports object to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[string]string{
		forever_constants.ContainerTypeLabel: forever_constants.ContainerType_EngineServer,
		PortSpecsLabel:                       serializedPortsSpec,
	}

	objectAttributes, err := newObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *objectAttributesProviderImpl) ForEnclave(enclaveId string) EnclaveObjectAttributesProvider {
	return newEnclaveObjectAttributesProviderImpl(enclaveId)
}
