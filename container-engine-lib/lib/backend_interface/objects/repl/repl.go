package repl

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type ReplGUID string

type Repl struct {
	guid ReplGUID
	enclaveId enclave.EnclaveID
}

func NewRepl(guid ReplGUID, enclaveId enclave.EnclaveID) *Repl {
	return &Repl{guid: guid, enclaveId: enclaveId}
}

func (repl *Repl) GetGUID() ReplGUID {
	return repl.guid
}

func (repl *Repl) GetEnclaveId() enclave.EnclaveID {
	return repl.enclaveId
}
