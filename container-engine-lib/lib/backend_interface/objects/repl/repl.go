package repl

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type ReplGUID string

// Object that represents POINT-IN-TIME information about an REPL
// Store this object and continue to reference it at your own risk!!!
type Repl struct {
	guid ReplGUID
	enclaveId enclave.EnclaveID
	status container_status.ContainerStatus
}

func NewRepl(guid ReplGUID, enclaveId enclave.EnclaveID, status container_status.ContainerStatus) *Repl {
	return &Repl{guid: guid, enclaveId: enclaveId, status: status}
}

func (repl *Repl) GetGUID() ReplGUID {
	return repl.guid
}

func (repl *Repl) GetEnclaveID() enclave.EnclaveID {
	return repl.enclaveId
}

func (repl *Repl) GetStatus() container_status.ContainerStatus {
	return repl.status
}
