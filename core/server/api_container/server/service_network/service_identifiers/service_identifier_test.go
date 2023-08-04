package service_identifiers

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServiceIdentifierMarshallers(t *testing.T) {
	serviceIdentifiersForTest := getServiceIdentifiersForTest()

	originalServiceIdentifier := serviceIdentifiersForTest[0]

	marshaledServiceIdentifier, err := json.Marshal(originalServiceIdentifier)
	require.NoError(t, err)
	require.NotNil(t, marshaledServiceIdentifier)

	newServiceIdentifier := &serviceIdentifier{}

	err = json.Unmarshal(marshaledServiceIdentifier, newServiceIdentifier)
	require.NoError(t, err)

	require.EqualValues(t, originalServiceIdentifier, newServiceIdentifier)
}
