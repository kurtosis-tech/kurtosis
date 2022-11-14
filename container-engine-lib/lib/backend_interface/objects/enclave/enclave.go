package enclave

import "time"

type EnclaveID string

type Enclave struct {
	id EnclaveID
	status EnclaveStatus
	creationTime *time.Time
}

func NewEnclave(id EnclaveID, status EnclaveStatus, creationTime *time.Time) *Enclave {
	return &Enclave{id: id, status: status, creationTime: creationTime}
}

func (enclave *Enclave) GetID() EnclaveID {
	return enclave.id
}

func (enclave *Enclave) GetStatus() EnclaveStatus {
	return enclave.status
}

func (enclave *Enclave) GetCreationTime() *time.Time {
	return enclave.creationTime
}
