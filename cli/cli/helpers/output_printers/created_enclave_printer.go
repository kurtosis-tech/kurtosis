package output_printers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
)

const createEnclaveTitle = "Created enclave"

func PrintEnclaveId(enclaveId enclaves.EnclaveID) {
	createdEnclaveMsg := fmt.Sprintf("%v: %v",createEnclaveTitle, enclaveId)
	GetFeaturedMessagePrinter().Print(createdEnclaveMsg)
}
