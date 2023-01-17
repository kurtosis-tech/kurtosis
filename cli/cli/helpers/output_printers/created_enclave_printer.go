package output_printers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
)

const createEnclaveTitle = "Created enclave"

func PrintEnclaveUUID(enclaveName string, enclaveUuid enclaves.EnclaveUUID) {
	createdEnclaveMsg := fmt.Sprintf("%v with name %v and UUID %v", createEnclaveTitle, enclaveName, enclaveUuid)
	GetSpotlightMessagePrinter().Print(createdEnclaveMsg)
}
