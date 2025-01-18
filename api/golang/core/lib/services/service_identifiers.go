package services

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	validUuidMatchesAllowed = 1
)

// Docs available at https://docs.kurtosis.com/sdk#service-identifiers
type ServiceIdentifiers struct {
	serviceNameToUuids          map[ServiceName][]ServiceUUID
	serviceUuids                map[ServiceUUID]bool
	serviceShortenedUuidToUuids map[string][]ServiceUUID
	enclaveNameForLogging       string
}

func NewServiceIdentifiers(enclaveNameForLogging string, allIdentifiers []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers) *ServiceIdentifiers {
	serviceNames := map[ServiceName][]ServiceUUID{}
	serviceUuids := map[ServiceUUID]bool{}
	serviceShortenedUuid := map[string][]ServiceUUID{}

	for _, serviceIdentifiers := range allIdentifiers {
		serviceName := ServiceName(serviceIdentifiers.GetName())
		serviceUuid := ServiceUUID(serviceIdentifiers.GetServiceUuid())
		serviceUuids[serviceUuid] = true
		shortenedUuid := serviceIdentifiers.GetShortenedUuid()
		serviceNames[serviceName] = append(serviceNames[serviceName], serviceUuid)
		serviceShortenedUuid[shortenedUuid] = append(serviceShortenedUuid[shortenedUuid], serviceUuid)
	}

	return &ServiceIdentifiers{enclaveNameForLogging: string(enclaveNameForLogging), serviceNameToUuids: serviceNames, serviceUuids: serviceUuids, serviceShortenedUuidToUuids: serviceShortenedUuid}
}

func (identifiers *ServiceIdentifiers) GetServiceUuidForIdentifier(identifier string) (ServiceUUID, error) {
	if _, found := identifiers.serviceUuids[ServiceUUID(identifier)]; found {
		return ServiceUUID(identifier), nil
	}

	if matches, found := identifiers.serviceShortenedUuidToUuids[identifier]; found {
		if len(matches) == validUuidMatchesAllowed {
			return matches[0], nil
		} else if len(matches) > validUuidMatchesAllowed {
			return "", stacktrace.NewError("Found multiple services '%v' matching shortened uuid '%v' in enclave '%v'. Please use a uuid to be more specific", matches, identifier, identifiers.enclaveNameForLogging)
		}
	}

	if matches, found := identifiers.serviceNameToUuids[ServiceName(identifier)]; found {
		if len(matches) == validUuidMatchesAllowed {
			return matches[0], nil
		} else if len(matches) > validUuidMatchesAllowed {
			return "", stacktrace.NewError("Found multiple services '%v' matching name '%v' in enclave '%v'. Please use a uuid to be more specific", matches, identifier, identifiers.enclaveNameForLogging)
		}
	}

	return "", stacktrace.NewError("No matching uuid for identifier '%s'", identifier)
}

func (identifiers *ServiceIdentifiers) GetOrderedListOfNames() []string {
	var serviceNames []string
	for name := range identifiers.serviceNameToUuids {
		serviceNames = append(serviceNames, string(name))
	}
	sort.Strings(serviceNames)
	return serviceNames
}
