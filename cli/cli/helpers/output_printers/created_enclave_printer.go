package output_printers

import (
	"fmt"
)

const createEnclaveTitle = "Created enclave:"

func PrintEnclaveName(enclaveName string) {
	createdEnclaveMsg := fmt.Sprintf("%v %v", createEnclaveTitle, enclaveName)
	GetSpotlightMessagePrinter().PrintWithLogger(createdEnclaveMsg)
}
