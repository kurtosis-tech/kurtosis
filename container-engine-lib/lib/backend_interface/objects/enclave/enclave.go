package enclave

type EnclaveID string

type Enclave struct {
	id EnclaveID
	status EnclaveStatus
}

func NewEnclave(id EnclaveID, status EnclaveStatus) *Enclave {
	return &Enclave{id: id, status: status}
}

func (enclave *Enclave) GetID() EnclaveID {
	return enclave.id
}

func (enclave *Enclave) GetStatus() EnclaveStatus {
	return enclave.status
}
