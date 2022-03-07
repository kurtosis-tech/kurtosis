package enclave

type Enclave struct {
	id string

	status EnclaveStatus
}
func (enclave *Enclave) GetID() string {
	return enclave.id
}
func (enclave *Enclave) GetStatus() EnclaveStatus {
	return enclave.status
}

