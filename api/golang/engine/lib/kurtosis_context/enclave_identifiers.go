package kurtosis_context

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

// Docs available at https://docs.kurtosis.com/sdk#enclave-identifiers
type EnclaveIdentifiers struct {
	enclaveNameToUuids          map[string][]enclaves.EnclaveUUID
	enclaveUuids                map[enclaves.EnclaveUUID]bool
	enclaveShortenedUuidToUuids map[string][]enclaves.EnclaveUUID
}

func newEnclaveIdentifiers(allIdentifiers []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers) *EnclaveIdentifiers {
	enclaveNames := map[string][]enclaves.EnclaveUUID{}
	enclaveUuids := map[enclaves.EnclaveUUID]bool{}
	enclaveShortenedUuid := map[string][]enclaves.EnclaveUUID{}
	for _, enclaveIdentifiers := range allIdentifiers {
		enclaveName := enclaveIdentifiers.GetName()
		enclaveUuid := enclaves.EnclaveUUID(enclaveIdentifiers.EnclaveUuid)
		enclaveUuids[enclaveUuid] = true
		shortenedUuid := enclaveIdentifiers.GetShortenedUuid()
		enclaveNames[enclaveName] = append(enclaveNames[enclaveName], enclaveUuid)
		enclaveShortenedUuid[shortenedUuid] = append(enclaveShortenedUuid[shortenedUuid], enclaveUuid)
	}

	return &EnclaveIdentifiers{enclaveNameToUuids: enclaveNames, enclaveUuids: enclaveUuids, enclaveShortenedUuidToUuids: enclaveShortenedUuid}
}

func (identifiers *EnclaveIdentifiers) GetEnclaveUuidForIdentifier(identifier string) (enclaves.EnclaveUUID, error) {
	if _, found := identifiers.enclaveUuids[enclaves.EnclaveUUID(identifier)]; found {
		return enclaves.EnclaveUUID(identifier), nil
	}

	if matches, found := identifiers.enclaveShortenedUuidToUuids[identifier]; found {
		if len(matches) == validUuidMatchesAllowed {
			return matches[0], nil
		} else if len(matches) > validUuidMatchesAllowed {
			return "", stacktrace.NewError("Found multiple enclaves '%v' matching shortened uuid '%v'. Please use a uuid to be more specific", matches, identifier)
		}
	}

	if matches, found := identifiers.enclaveNameToUuids[identifier]; found {
		if len(matches) == validUuidMatchesAllowed {
			return matches[0], nil
		} else if len(matches) > validUuidMatchesAllowed {
			return "", stacktrace.NewError("Found multiple enclaves '%v' matching name '%v'. Please use a uuid to be more specific", matches, identifier)
		}
	}

	return "", stacktrace.NewError("No matching uuid for identifier '%s'", identifier)
}

func (identifiers *EnclaveIdentifiers) GetOrderedListOfNames() []string {
	var enclaveNames []string

	for name := range identifiers.enclaveNameToUuids {
		enclaveNames = append(enclaveNames, name)
	}

	sort.Strings(enclaveNames)
	return enclaveNames
}
