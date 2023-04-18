package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testBarService = service.ServiceName("bar")
	fooPortId      = "foo"
	fizzPortId     = "fizz"
	invalidPortId  = "invalid"
)

func TestMultiplePortIdsForValidation(t *testing.T) {
	validatorEnvironment := NewValidatorEnvironment(false, nil, nil)
	validatorEnvironment.AddPrivatePortIDForService(fooPortId, testBarService)
	validatorEnvironment.AddPrivatePortIDForService(fizzPortId, testBarService)
	require.True(t, validatorEnvironment.DoesPrivatePortIDExistForService(fooPortId, testBarService))
	require.True(t, validatorEnvironment.DoesPrivatePortIDExistForService(fizzPortId, testBarService))
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(invalidPortId, testBarService))
	validatorEnvironment.RemoveServiceFromPrivatePortIDMapping(testBarService)
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(fooPortId, testBarService))
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(fizzPortId, testBarService))
}
