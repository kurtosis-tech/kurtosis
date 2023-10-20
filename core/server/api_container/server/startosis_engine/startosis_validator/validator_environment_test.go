package startosis_validator

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
)

const (
	testBarService                = service.ServiceName("bar")
	fooPortId                     = "foo"
	fizzPortId                    = "fizz"
	invalidPortId                 = "invalid"
	availableMemoryInBytes        = 12000
	availableCpuInMilliCores      = 4231
	isResourceInformationComplete = true
	tooMuchMemory                 = 120000
	tooMuchCpu                    = 5000
)

func TestMultiplePortIdsForValidation(t *testing.T) {
	emptyInitialMapping := map[service.ServiceName][]string{}
	validatorEnvironment := NewValidatorEnvironment(nil, nil, emptyInitialMapping, availableCpuInMilliCores, availableMemoryInBytes, isResourceInformationComplete, image_download_mode.ImageDownloadMode_Missing)
	portIds := []string{
		fooPortId,
		fizzPortId,
	}
	validatorEnvironment.AddPrivatePortIDForService(portIds, testBarService)
	require.True(t, validatorEnvironment.DoesPrivatePortIDExistForService(fooPortId, testBarService))
	require.True(t, validatorEnvironment.DoesPrivatePortIDExistForService(fizzPortId, testBarService))
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(invalidPortId, testBarService))
	validatorEnvironment.RemoveServiceFromPrivatePortIDMapping(testBarService)
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(fooPortId, testBarService))
	require.False(t, validatorEnvironment.DoesPrivatePortIDExistForService(fizzPortId, testBarService))
	require.Error(t, validatorEnvironment.HasEnoughCPU(tooMuchCpu, testBarService))
	require.Error(t, validatorEnvironment.HasEnoughMemory(tooMuchMemory, testBarService))
}
