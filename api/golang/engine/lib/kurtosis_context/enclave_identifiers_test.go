package kurtosis_context

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	firstUuid  = "fe5a7d34f4424bd49fa265983def24e7"
	secondUuid = "39cbe546b1e249adb5286bee0c844e5c"
	thirdUuid  = "5210fe194d3246fd9eabb6a9928796d2"

	firstEnclaveName  = "best-enclave-one"
	secondEnclaveName = "enclave-two"
	thirdEnclaveName  = "enclave-three"

	shortenedUuidLength = 12
)

var (
	firstShortenedUuid  = firstUuid[:shortenedUuidLength]
	secondShortenedUuid = secondUuid[:shortenedUuidLength]
	thirdShortenedUuid  = thirdUuid[:shortenedUuidLength]

	firstEnclaveIdentifiers = &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
		EnclaveUuid:   firstUuid,
		Name:          firstEnclaveName,
		ShortenedUuid: firstShortenedUuid,
	}

	secondEnclaveIdentifiers = &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
		EnclaveUuid:   secondUuid,
		Name:          secondEnclaveName,
		ShortenedUuid: secondShortenedUuid,
	}

	thirdEnclaveIdentifiers = &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
		EnclaveUuid:   thirdUuid,
		Name:          thirdEnclaveName,
		ShortenedUuid: thirdShortenedUuid,
	}

	combinedEnclaveIdentifiers = []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{firstEnclaveIdentifiers, secondEnclaveIdentifiers, thirdEnclaveIdentifiers}
)

func TestEnclaveIdentifiers_GetUuidForIdentifier(t *testing.T) {
	testEnclaveIdentifiers := newEnclaveIdentifiers(combinedEnclaveIdentifiers)
	require.NotNil(t, testEnclaveIdentifiers)

	for _, enclaveIdentifier := range combinedEnclaveIdentifiers {
		uuidByUuid, err := testEnclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveIdentifier.EnclaveUuid)
		require.Nil(t, err)
		uuidByName, err := testEnclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveIdentifier.Name)
		require.Nil(t, err)
		uuidByShortenedUuid, err := testEnclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveIdentifier.ShortenedUuid)
		require.Nil(t, err)

		require.Equal(t, uuidByUuid, uuidByName)
		require.Equal(t, uuidByUuid, uuidByShortenedUuid)
		require.Equal(t, enclaveIdentifier.EnclaveUuid, string(uuidByName))
	}
}

func TestEnclaveIdentifiers_OrderedNames(t *testing.T) {
	testEnclaveIdentifiers := newEnclaveIdentifiers(combinedEnclaveIdentifiers)
	require.NotNil(t, testEnclaveIdentifiers)

	expectedOrder := []string{
		firstEnclaveName,
		thirdEnclaveName,
		secondEnclaveName,
	}

	require.Equal(t, expectedOrder, testEnclaveIdentifiers.GetOrderedListOfNames())
}

func TestEnclaveIdentifiers_GetUuidForIdentifierFailureModes(t *testing.T) {
	dupeIdentifiers := append(combinedEnclaveIdentifiers, firstEnclaveIdentifiers)
	testEnclaveIdentifiers := newEnclaveIdentifiers(dupeIdentifiers)

	// dupe valid name
	_, err := testEnclaveIdentifiers.GetEnclaveUuidForIdentifier(firstEnclaveName)
	require.Error(t, err)

	// dupe shortened uuid
	_, err = testEnclaveIdentifiers.GetEnclaveUuidForIdentifier(firstShortenedUuid)
	require.Error(t, err)

	// invalid name
	_, err = testEnclaveIdentifiers.GetEnclaveUuidForIdentifier("invalid-identifier")
	require.Error(t, err)
}
